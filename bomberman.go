package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	//"runtime"
	"time"

	//"github.com/aybabtme/bombertcp"
	"github.com/nsf/termbox-go"

	"github.com/setnicka/bomberman/board"
	"github.com/setnicka/bomberman/game"
	"github.com/setnicka/bomberman/logger"
	"github.com/setnicka/bomberman/objects"
	"github.com/setnicka/bomberman/player"
	"github.com/setnicka/bomberman/player/ai"
	"github.com/setnicka/bomberman/player/input"
	"github.com/setnicka/bomberman/player/websocket"
	"github.com/setnicka/bomberman/scheduler"
)

func init() {
	websocketplayer.SetLog(log)
}

const LogLevel = logger.Debug

var (
	log = logger.New("", "bomb.log", LogLevel)

	w, h          int
	config        Config
	playersConfig PlayersConf
	players       = map[string]player.Player{}
	publicWatcher *websocketplayer.Player
	pointsFile    string
	gameId        string
)

func main() {
	log.Infof("Starting Bomberman")
	rand.Seed(time.Now().UTC().UnixNano())

	// 1. Parse command line arguments
	var (
		gameConfigFile    string
		playersConfigFile string
		port              int
		debug             bool
		drawMap           bool
	)
	flag.StringVar(&gameConfigFile, "config", "config.json", "Choose `file` with game configuration.")
	flag.StringVar(&playersConfigFile, "players", "players.json", "Choose `file` with players configuration.")
	flag.IntVar(&port, "port", 8000, "Set `port` for remote players")
	flag.BoolVar(&debug, "debug", false, "Enable control of some player from keyboard")
	flag.BoolVar(&drawMap, "map", false, "Enable drawing map in the console")
	flag.StringVar(&pointsFile, "points", "points.json", "Choose `file` to which the points will be saved.")
	flag.StringVar(&gameId, "id", time.Now().Format("2006-01-02_15:04:05"), "Set game `identificator` for saving into points file")
	flag.Parse()

	// 2. Load config
	rawConfig, err := ioutil.ReadFile(gameConfigFile)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(rawConfig, &config); err != nil {
		panic(err)
	}

	// 2. Load players
	rawPlayers, err := ioutil.ReadFile(playersConfigFile)
	if err != nil {
		panic(err)
	}
	if err := json.Unmarshal(rawPlayers, &playersConfig); err != nil {
		panic(err)
	}
	randomPositions := genRandomPositions(len(playersConfig))
	for i, player := range playersConfig {
		switch player.Position {
		case "random":
			player.StartX = rand.Intn(config.Width) + 1
			player.StartY = rand.Intn(config.Height) + 1
		case "randomList":
			player.StartX, player.StartY = randomPositions[i][0], randomPositions[i][1]
		default:
			if player.StartX < 0 {
				player.StartX += config.Width + 1
			}
			if player.StartY < 0 {
				player.StartY += config.Height + 1
			}
		}
		playersConfig[i] = player
		log.Debugf("Player '%s' starts at '%dx%d'", player.Name, player.StartX, player.StartY)
	}
	log.Debugf("Players: %+v", playersConfig)

	// 3. init game
	log.Infof("Initializing game")
	turnDuration := time.Duration(config.TurnDuration) * time.Millisecond
	game := game.NewGame(turnDuration, config.TotalBombsPowerups, config.TotalRadiusPowerups)

	game.Players = map[*player.State]player.Player{}
	var inputChan chan player.Move

	for i, p := range playersConfig {
		state := player.State{
			BasicState: player.BasicState{
				Name:    p.Name,
				Number:  i,
				X:       p.StartX,
				Y:       p.StartY,
				LastX:   -1,
				LastY:   -1,
				Bombs:   config.DefaultMaxBombs,
				MaxBomb: config.DefaultMaxBombs,
				Radius:  config.DefaultBombRadius,
				Alive:   true,
			},
			Type:       p.Type,
			Symbol:     p.Symbol,
			GameObject: &objects.TboxPlayer{p.Symbol},
		}

		switch p.Type {
		case LOCAL_PLAYER:
			if inputChan != nil {
				panic(fmt.Errorf("Cannot have more than one local player, player '%s' could not be initialized", p.Name))
			}
			inputChan = make(chan player.Move)
			players[p.ID] = inputplayer.New(state, inputChan)
		case WEBSOCKET_PLAYER:
			players[p.ID] = websocketplayer.New(&state)
		case AI_PLAYER:
			players[p.ID] = ai.NewRandomPlayer(state, int64(i))
		}
		game.Players[&state] = players[p.ID]
		//game.Players[&state] = bombertcp.NewTcpPlayer(state, "0.0.0.0:40000", log)
	}
	// Add dead public player for watching the game
	state := player.State{BasicState: player.BasicState{Alive: false}, Hidden: true}
	publicWatcher = websocketplayer.New(&state)
	game.Players[&state] = publicWatcher

	//runtime.GOMAXPROCS(1 + 2*len(game.Players))

	log.Debugf("Setup board.")
	board := board.SetupBoard(game, config.Width+2, config.Height+2, config.FreeAreaAroundPlayers, config.RockDensity)
	exportedBoard := board.Export()
	// Construct links for basic state of other players
	playersStates := []*player.BasicState{}
	for pState := range game.Players {
		if !pState.Hidden {
			playersStates = append(playersStates, &pState.BasicState)
		}
	}
	// Add links to other players and to the current exported board to the players states
	for pState := range game.Players {
		pState.Board = &exportedBoard
		pState.Players = playersStates
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
			if pm, ok := toPlayerMove(ev); ok && inputChan != nil {
				select {
				case inputChan <- pm:
				default:
				}
			} else {
				evChan <- ev
			}
		}
	}()

	log.Debugf("Drawing for first time.")
	consoleDraw(board, game.Players, drawMap)

	log.Infof("Starting in %d seconds.", config.StartCountdown)
	startTimer := time.NewTimer(time.Duration(config.StartCountdown) * time.Second)
outer:
	for {
		select {
		case <-game.TurnTick.C:
			consoleDraw(board, game.Players, drawMap)
		case <-startTimer.C:
			break outer
		}
	}
	log.Infof("Starting.")

	MainLoop(game, board, evChan, drawMap)
}

func MainLoop(g *game.Game, board board.Board, evChan <-chan termbox.Event, drawMap bool) {
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

		updatePlayers(g, board)
		savePoints(g.Players)
		consoleDraw(board, g.Players, drawMap)

		alives := []player.Player{}
		for pState, player := range g.Players {
			if pState.Alive {
				alives = append(alives, player)
			}
		}
		if len(alives) == 1 && config.AutoStopGame {
			log.Infof("%s won. All other players are dead.", alives[0].Name())
			return
		} else if len(alives) == 0 && config.AutoStopGame {
			log.Infof("Draw! All players are dead.")
			return
		}
	}
}

func genRandomPositions(n int) [][2]int {
	// Random positions - for max 4 players
	w := config.Width
	h := config.Height
	positions := [][2]int{{1, 1}, {1, h}, {w, 1}, {w, h}}
	// If there are more than 4 players add middles of sides
	if len(playersConfig) > 4 {
		positions = append(positions, [][2]int{
			{1, h / 2}, {w / 2, 1}, {w, h / 2}, {w / 2, h},
		}...)
	}
	if len(playersConfig) > 8 {
		// Add four positions in smaller square near the middle
		positions = append(positions, [][2]int{
			{w / 3, h / 3}, {w / 3, 2 * h / 3}, {2 * w / 3, h / 3}, {2 * w / 3, 2 * h / 3},
		}...)
	}

	// random shuffle and return
	rand.Shuffle(n, func(i, j int) {
		positions[i], positions[j] = positions[j], positions[i]
	})
	return positions
}

func savePoints(players map[*player.State]player.Player) {
	results := PointResults{}

	// 1. Load points file
	if oldContent, err := ioutil.ReadFile(pointsFile); err == nil {
		if err := json.Unmarshal(oldContent, &results); err != nil {
			log.Errorf("Cannot unmarshal points from the points file: %v", err)
		}
	}

	// 2. Set values
	change := false
	results[gameId] = map[string]int{}
	for state := range players {
		if !state.Hidden {
			if results[gameId][state.Name] != state.Points {
				change = true
			}
			results[gameId][state.Name] = state.Points
		}
	}

	// 3. Compute total points
	totalPoints := map[string]int{}
	for _, gameResults := range results {
		for playerName, points := range gameResults {
			totalPoints[playerName] += points
		}
	}
	for state := range players {
		if value, found := totalPoints[state.Name]; found {
			state.TotalPoints = value
		}
	}

	// When no change no need to save points
	if !change {
		return
	}

	// 4. Save into file
	newContent, err := json.Marshal(results)
	if err != nil {
		log.Errorf("Cannot marshal results: %v", err)
		return
	}
	if err := ioutil.WriteFile(pointsFile, newContent, 0644); err != nil {
		log.Errorf("Cannot save results: %v", err)
	}
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
	exportedBoard := board.Export()
	for pState, player := range game.Players {
		pState.Board = &exportedBoard
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
		if pState.Alive && board[nextX][nextY].Top() == objects.Flame {
			pState.Alive = false
			log.Infof("[%s] Died moving into flame.", pState.Name)
			// Count points for such obvious suicide
			log.Infof("[%s] Receiving %d points for suicide", pState.Name, config.PointsPerSuicide)
			pState.Points += config.PointsPerSuicide
			// Remove player from cell
			cell := board[pState.X][pState.Y]
			if !cell.Remove(pState.GameObject) {
				log.Panicf("[%s] player not found at (%d, %d), cell=%#v",
					pState.Name, pState.X, pState.Y, cell)
			}
		}
		if !pState.Alive {
			// Player was removed when was killed
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
		pState.Bombs++
		c.Pop()
		log.Infof("[%s] Powerup! Max bombs: %d", pState.Name, pState.MaxBomb)
	case objects.RadiusPU:
		pState.Radius++
		c.Pop()
		log.Infof("[%s] Powerup! Radius: %d", pState.Name, pState.Radius)
	}
}
