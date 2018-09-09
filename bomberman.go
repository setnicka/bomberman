package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"runtime"
	"time"

	//"github.com/aybabtme/bombertcp"
	"github.com/nsf/termbox-go"

	"github.com/setnicka/bomberman/board"
	"github.com/setnicka/bomberman/game"
	"github.com/setnicka/bomberman/logger"
	"github.com/setnicka/bomberman/objects"
	"github.com/setnicka/bomberman/player"
	"github.com/setnicka/bomberman/player/input"
	"github.com/setnicka/bomberman/scheduler"
)

const LogLevel = logger.Debug

var (
	log = logger.New("", "bomb.log", LogLevel)

	w, h          int
	config        Config
	playersConfig PlayersConf
	remotePlayers = map[string]*RemotePlayer{}
	publicWatcher *RemotePlayer
)

func main() {
	log.Infof("Starting Bomberman")
	rand.Seed(time.Now().UTC().UnixNano())

	// 1. Parse command line arguments
	var (
		gameConfigFile    string
		playersConfigFile string
		port              int
	)
	flag.StringVar(&gameConfigFile, "config", "config.json", "Choose `file` with game configuration.")
	flag.StringVar(&playersConfigFile, "players", "players.json", "Choose `file` with players configuration.")
	flag.IntVar(&port, "port", 8000, "Set `port` for remote players")
	flag.Parse()

	// 2. Load config
	rawConfig, err := ioutil.ReadFile(gameConfigFile)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		panic(err)
	}
	log.Debugf("Config: %+v", config)

	// 2. Load players
	rawPlayers, err := ioutil.ReadFile(playersConfigFile)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(rawPlayers, &playersConfig); err != nil {
		panic(err)
	}
	for i, player := range playersConfig {
		if player.StartX < 0 {
			player.StartX += config.Width + 1
		}
		if player.StartY < 0 {
			player.StartY += config.Height + 1
		}
		playersConfig[i] = player
	}
	log.Debugf("Players: %+v", playersConfig)

	// 3. init game
	log.Infof("Initializing game")
	turnDuration := time.Duration(config.TurnDuration) * time.Millisecond
	game := game.NewGame(turnDuration, config.TotalBombsPowerups, config.TotalRadiusPowerups)

	game.Players = map[*player.State]player.Player{}
	for _, p := range playersConfig {
		state := player.State{
			Name:         p.Name,
			X:            p.StartX,
			Y:            p.StartY,
			LastX:        -1,
			LastY:        -1,
			TurnDuration: turnDuration,
			Bombs:        0,
			MaxBomb:      config.DefaultMaxBombs,
			MaxRadius:    config.DefaultBombRadius,
			Alive:        true,
			GameObject:   &objects.TboxPlayer{p.Symbol},
		}
		//game.Players[&state] = bombertcp.NewTcpPlayer(state, "0.0.0.0:40000", log)
		remotePlayers[p.ID] = NewRemotePlayer(p, state)
		game.Players[&state] = remotePlayers[p.ID]
	}
	// Add dead public player for watching the game
	state := player.State{Alive: false}
	publicWatcher = NewRemotePlayer(PlayerConf{Name: "Public"}, state)
	game.Players[&state] = publicWatcher

	runtime.GOMAXPROCS(1 + len(game.Players))

	log.Debugf("Setup board.")
	board := board.SetupBoard(game, config.Width+2, config.Height+2, config.FreeAreaAroundPlayers, config.RockDensity)
	for pState := range game.Players {
		pState.Board = board.Clone()
	}

	// 4. Init WebSockets connection
	go WebsocketsStart(port)

	// 5. Terminal initialization
	log.Debugf("Initializing termbox.")
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	w, h = termbox.Size()

	log.Debugf("Initializing termbox event poller.")
	evChan := make(chan termbox.Event)
	go func() {
		log.Debugf("Polling events.")
		for {
			ev := termbox.PollEvent()
			/*
				if pm, ok := toPlayerMove(ev); ok {
					select {
					case inputChan <- pm:
					default:
						log.Debugf("Dropping event '%#v', player not reading.", ev.Type)
					}
				} else {
					evChan <- ev
				}
			*/
			evChan <- ev
		}
	}()

	log.Debugf("Drawing for first time.")
	board.Draw(game.Players)

	log.Debugf("Starting.")

	MainLoop(game, board, evChan)
}

func MainLoop(g *game.Game, board board.Board, evChan <-chan termbox.Event) {
	for range g.TurnTick.C {
		if g.IsDone() {
			log.Infof("Game requested to stop.")
			return
		}

		receiveEvents(g, evChan)

		applyPlayerMoves(g, board)

		g.RunSchedule(func(a scheduler.Action, turn int) error {
			act := a.(*BomberAction)
			log.Debugf("[%s] !!! turn %d/%d", act.name, turn, act.Duration())
			return act.doTurn(turn)
		})

		board.Draw(g.Players)
		updatePlayers(g, board)

		alives := []player.Player{}
		for pState, player := range g.Players {
			if pState.Alive {
				alives = append(alives, player)
			}
		}
		if len(alives) == 1 {
			log.Infof("%s won. All other players are dead.", alives[0].Name())
			return
		} else if len(alives) == 0 {
			log.Infof("Draw! All players are dead.")
			return
		}
	}
}

func initLocalPlayer(pState player.State) (player.Player, chan<- player.Move) {
	keyPlayerChan := make(chan player.Move, 1)
	keyPlayer := input.NewInputPlayer(pState, keyPlayerChan)
	return keyPlayer, keyPlayerChan
}

//////////////
// Events

func receiveEvents(g *game.Game, evChan <-chan termbox.Event) {
	select {
	case ev := <-evChan:
		switch ev.Type {
		case termbox.EventResize:
			w, h = ev.Width, ev.Height
		case termbox.EventError:
			g.SetDone()
		case termbox.EventKey:
			doKey(g, ev.Key)
		}
	default:
	}
}

func doKey(g *game.Game, key termbox.Key) {
	switch key {
	case termbox.KeyCtrlC:
		g.SetDone()
	}
}

//////////////
// Schedule

type BomberAction struct {
	name     string
	duration int
	doTurn   func(turn int) error
}

func (a *BomberAction) Duration() int {
	return a.duration
}

//////////////
// Players

func applyPlayerMoves(g *game.Game, board board.Board) {
	for pState, player := range g.Players {
		if pState.Alive {
			select {
			case m := <-player.Move():
				movePlayer(g, board, pState, m)
			default:
			}
		}
	}
}

func updatePlayers(game *game.Game, board board.Board) {
	for pState, player := range game.Players {
		pState.Board = board.Clone()
		pState.Turn = game.Turn()
		select {
		case player.Update() <- *pState:
		default:
		}
	}
}

func toPlayerMove(ev termbox.Event) (player.Move, bool) {
	if ev.Type != termbox.EventKey {
		return player.Move(""), false
	}

	switch ev.Key {
	case termbox.KeyArrowUp:
		return player.Up, true
	case termbox.KeyArrowDown:
		return player.Down, true
	case termbox.KeyArrowLeft:
		return player.Left, true
	case termbox.KeyArrowRight:
		return player.Right, true
	case termbox.KeySpace:
		return player.PutBomb, true
	}

	return player.Move(""), false
}

func movePlayer(g *game.Game, board board.Board, pState *player.State, action player.Move) {
	nextX, nextY := pState.X, pState.Y
	switch action {
	case player.Up:
		nextY--
	case player.Down:
		nextY++
	case player.Left:
		nextX--
	case player.Right:
		nextX++
	case player.PutBomb:
		placeBomb(board, g, pState)
	}

	if !board.Traversable(nextX, nextY) {
		return
	}

	doMove := func(turn int) error {
		if board[nextX][nextY].Top() == objects.Flame {
			pState.Alive = false
			log.Infof("[%s] Died moving into flame.", pState.Name)
			cell := board[pState.X][pState.Y]
			if !cell.Remove(pState.GameObject) {
				log.Panicf("[%s] player not found at (%d, %d), cell=%#v",
					pState.Name, pState.X, pState.Y, cell)
			}
			return nil
		}

		pState.LastX, pState.LastY = pState.X, pState.Y
		pState.X, pState.Y = nextX, nextY

		pickPowerUps(board, pState, nextX, nextY)

		cell := board[pState.LastX][pState.LastY]
		if !cell.Remove(pState.GameObject) {
			log.Panicf("[%s] player not found at (%d, %d), cell=%#v",
				pState.Name, pState.X, pState.Y, cell)
		}
		board[nextX][nextY].Push(pState.GameObject)

		return nil
	}

	g.Schedule.Register(&BomberAction{
		name:     fmt.Sprintf("%s.moving(%#v)", pState.Name, action),
		duration: 1,
		doTurn:   doMove,
	}, 1)

}

func pickPowerUps(board board.Board, pState *player.State, x, y int) {
	c := board[x][y]
	switch c.Top() {
	case objects.BombPU:
		pState.MaxBomb++
		c.Pop()
		log.Infof("[%s] Powerup! Max bombs: %d", pState.Name, pState.MaxBomb)
	case objects.RadiusPU:
		pState.MaxRadius++
		c.Pop()
		log.Infof("[%s] Powerup! Max radius: %d", pState.Name, pState.MaxRadius)
	}
}
