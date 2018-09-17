package websocketplayer

import (
	"encoding/json"

	"github.com/gorilla/websocket"

	"github.com/setnicka/bomberman/player"
)

type Client struct {
	Id         int
	Player     *Player
	Conn       *websocket.Conn
	updateChan chan *player.State
	moveChan   chan player.Move
}

// Send game state each turn
func (c *Client) SendState() {
	playerName := c.Player.state.Name
	for {
		// Get state update
		update := <-c.updateChan
		log.Debugf("[Client %s:%d] Sending state update %d", playerName, c.Id, update.Turn)

		// Marshal into byte array
		msg, err := json.Marshal(update)
		if err != nil {
			log.Errorf("[Client %s:%d] Cannot marshal game state: %v", playerName, c.Id, err)
			c.Close()
			return
		}

		if err := c.Conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			log.Errorf("[Client %s:%d] Cannot send state update, closing connection: %v", playerName, c.Id, err)
			c.Close()
			return
		}
	}
}

func (c *Client) ReceiveMoves() {
	playerName := c.Player.state.Name
	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			log.Errorf("[Client %s:%d] Cannot receive move: %v", playerName, c.Id, err)
			c.Close()
			return
		}
		log.Debugf("[Client %s:%d] Received message: %s", playerName, c.Id, msg)
		c.moveChan <- player.Move(msg)
	}
}

func (c *Client) Close() {
	c.Conn.Close()
	// Unregister Client from WebsocketPlayer object
	log.Debugf("Stopping client %d for player '%s'", c.Id, c.Player.state.Name)
	c.Player.UnregisterClient(c.Id)
}
