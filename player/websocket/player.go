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
	state player.State

	// Outer communication
	updateChan  chan player.State
	moveChan    chan player.Move
	outMoveChan chan player.Move

	// State distribution to clients
	clientChannels       map[*Client]chan *player.State
	clientUnregisterChan chan *Client
	lastClientId         int

	responseTime time.Duration
}

func New(state player.State) *Player {
	p := &Player{
		state:                state,
		updateChan:           make(chan player.State),
		moveChan:             make(chan player.Move),
		outMoveChan:          make(chan player.Move, 1),
		clientChannels:       map[*Client]chan *player.State{},
		clientUnregisterChan: make(chan *Client),
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
			for _, clientChan := range p.clientChannels {
				clientChan <- &update
			}
		case unregisterClient := <-p.clientUnregisterChan:
			delete(p.clientChannels, unregisterClient)
		case move := <-p.moveChan:
			select {
			case p.outMoveChan <- move:
				p.responseTime = time.Since(sendTime)
				log.Debugf("[Player %s] Response time: %v", p.state.Name, p.responseTime)
			default:
				// skip all other moves in this round
			}
		}
	}
}

func (p *Player) StartClient(conn *websocket.Conn, gameSettings interface{}) {
	p.lastClientId++
	client := &Client{
		Id:         p.lastClientId,
		Player:     p,
		Conn:       conn,
		updateChan: make(chan *player.State, 1),
		moveChan:   p.moveChan,
	}

	p.clientChannels[client] = client.updateChan

	// Firstly send game settings
	msg, err := json.Marshal(gameSettings)
	if err != nil {
		log.Errorf("[Client %s:%d] Cannot marshal game settings: %v", p.state.Name, client.Id, err)
		client.Close()
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
		log.Errorf("[Client %s:%d] Cannot send game settings: %v", p.state.Name, client.Id, err)
		client.Close()
		return
	}

	// Then start workers
	go client.SendState()
	go client.ReceiveMoves()
}

func (p *Player) UnregisterClient(client *Client) {
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
