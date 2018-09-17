package websocketplayer

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"

	"github.com/setnicka/bomberman/logger"
	"github.com/setnicka/bomberman/player"
)

var log *logger.Logger

func SetLog(l *logger.Logger) {
	log = l
}

type Player struct {
	state *player.State

	// Outer communication
	updateChan  chan player.State
	moveChan    chan player.Move
	outMoveChan chan player.Move

	// State distribution to clients
	clientChannels       map[int]chan *player.State
	clientUnregisterChan chan int
	clientRegisterChan   chan *websocket.Conn
	lastClientId         int
}

func New(state *player.State) *Player {
	p := &Player{
		state:                state,
		updateChan:           make(chan player.State),
		moveChan:             make(chan player.Move),
		outMoveChan:          make(chan player.Move, 1),
		clientChannels:       map[int]chan *player.State{},
		clientUnregisterChan: make(chan int),
		clientRegisterChan:   make(chan *websocket.Conn),
	}

	go p.loop()

	return p
}

func (p *Player) loop() {
	var sendTime time.Time
	for {
		select {
		case update := <-p.updateChan:
			sendTime = time.Now()
			// Distribute update to all clients
			if len(p.clientChannels) > 0 {
				log.Debugf("SENDING INFO TO ALL CLIENTS OF %s", p.Name())
			}
			for _, clientChan := range p.clientChannels {
				clientChan <- &update
			}
		case unregisterClient := <-p.clientUnregisterChan:
			delete(p.clientChannels, unregisterClient)
			if len(p.clientChannels) == 0 {
				p.state.Connected = false
			}
		case move := <-p.moveChan:
			select {
			case p.outMoveChan <- move:
				p.state.ResponseTime = time.Since(sendTime)
				log.Debugf("[Player %s] Response time: %v", p.state.Name, p.state.ResponseTime)
			default:
				// skip all other moves in this round
			}
		case conn := <-p.clientRegisterChan:
			p.state.Connected = true
			p.lastClientId++
			client := &Client{
				Id:         p.lastClientId,
				Player:     p,
				Conn:       conn,
				updateChan: make(chan *player.State, 1),
				moveChan:   p.moveChan,
			}

			p.clientChannels[client.Id] = client.updateChan

			// Then start workers
			go client.SendState()
			go client.ReceiveMoves()
		}
	}
}

func (p *Player) StartClient(conn *websocket.Conn, gameSettings interface{}) {
	// Firstly send game settings
	msg, err := json.Marshal(gameSettings)
	if err != nil {
		log.Errorf("[Client %s:NEW] Cannot marshal game settings: %v", p.state.Name, err)
		conn.Close()
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Errorf("[Client %s:NEW] Cannot send game settings: %v", p.state.Name, err)
		conn.Close()
		return
	}

	p.clientRegisterChan <- conn

}

func (p *Player) UnregisterClient(client int) {
	p.clientUnregisterChan <- client
}

func (p *Player) Name() string {
	name := p.state.Name
	return name
}

func (p *Player) Move() <-chan player.Move {
	return p.outMoveChan
}
func (p *Player) Update() chan<- player.State {
	return p.updateChan
}
