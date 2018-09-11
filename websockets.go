package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"github.com/setnicka/bomberman/player"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func WebsocketsStart(port int) error {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// 1. Upgrade to websockets connection
		upgrader.CheckOrigin = func(r *http.Request) bool { return true }
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Errorf("Problem during upgrading to websockets: %v", err)
			return
		}

		// 2. Receive first message with id:password of with string "public"
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Errorf("Problem during receiving initial message: %v", err)
		}
		parts := strings.Split(string(msg), ":")
		if len(parts) == 2 {
			// Get player
			for _, player := range playersConfig {
				if player.ID == parts[0] && player.Password == parts[1] {
					log.Debugf("Logged in player '%s'", parts)

					remotePlayers[player.ID].StartClient(conn)
				}
			}
		} else {
			log.Debugf("Logged in public client")
			publicWatcher.StartClient(conn)
		}

		//
	})
	address := fmt.Sprintf(":%d", port)
	log.Infof("Starting websockets server on address '%s'", address)
	return http.ListenAndServe(address, nil)
}

func listenToClient(player *PlayerConf, w http.ResponseWriter, r *http.Request) {

}

// Send game state each turn
func (c *RemotePlayerClient) SendState() {
	playerConf := c.RemotePlayer.playerConf
	for {
		// Get state update
		update := <-c.updateChan
		log.Debugf("[Client %s:%d] Sending state update %d", playerConf.Name, c.Id, update.Turn)

		// Marshal into byte array
		msg, err := json.Marshal(update)
		if err != nil {
			log.Errorf("[Client %s:%d] Cannot marshal game state: %v", playerConf.Name, c.Id, err)
			c.Close()
			return
		}

		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Errorf("[Client %s:%d] Cannot send state update, closing connection: %v", playerConf.Name, c.Id, err)
			c.Close()
			return
		}
	}
}

func (c *RemotePlayerClient) ReceiveMoves() {
	playerConf := c.RemotePlayer.playerConf
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Errorf("[Client %s:%d] Cannot receive move: %v", playerConf.Name, c.Id, err)
			c.Close()
			return
		}
		log.Debugf("[Client %s:%d] Received message: %s", playerConf.Name, c.Id, msg)
		c.moveChan <- player.Move(msg)
	}
}

func (c *RemotePlayerClient) Close() {
	c.Conn.Close()
	// Unregister Client from remotePlayer object
	log.Debugf("Stopping client %d for player '%s'", c.Id, c.RemotePlayer.playerConf.Name)
	c.RemotePlayer.UnregisterClient(c)
}

///////////////////////////////////////

func NewRemotePlayer(playerConf PlayerConf, state player.State) *RemotePlayer {
	p := &RemotePlayer{
		playerConf:           playerConf,
		state:                state,
		updateChan:           make(chan player.State),
		moveChan:             make(chan player.Move),
		outMoveChan:          make(chan player.Move, 1),
		clientChannels:       map[*RemotePlayerClient]chan *player.State{},
		clientUnregisterChan: make(chan *RemotePlayerClient),
	}

	go p.loop()

	return p
}

func (p *RemotePlayer) loop() {
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
				log.Debugf("[Player %s] Response time: %v", p.playerConf.ID, p.responseTime)
			default:
				// skip all other moves in this round
			}
		}
	}
}

func (p *RemotePlayer) StartClient(conn *websocket.Conn) {
	p.lastClientId++
	client := &RemotePlayerClient{
		Id:           p.lastClientId,
		RemotePlayer: p,
		Conn:         conn,
		updateChan:   make(chan *player.State, 1),
		moveChan:     p.moveChan,
	}

	p.clientChannels[client] = client.updateChan

	go client.SendState()
	go client.ReceiveMoves()
}

func (p *RemotePlayer) UnregisterClient(client *RemotePlayerClient) {
	p.clientUnregisterChan <- client
}

func (p *RemotePlayer) Name() string {
	name := p.state.Name
	return name
}

func (p *RemotePlayer) Move() <-chan player.Move {
	return p.outMoveChan
}
func (p *RemotePlayer) Update() chan<- player.State {
	return p.updateChan
}

// For testing only:
func (p *RemotePlayer) forwardMove(move player.Move) {
	p.outMoveChan <- move
}
