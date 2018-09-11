package main

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"github.com/setnicka/bomberman/player/websocket"
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
					if wPlayer, ok := players[player.ID].(*websocketplayer.Player); ok {
						log.Debugf("Logged in player '%s'", parts)
						wPlayer.StartClient(conn, config.PublicConfig)
					} else {
						log.Debugf("Player '%s' is not websocket player! Not starting client.")
					}
				}
			}
		} else {
			log.Debugf("Logged in public client")
			publicWatcher.StartClient(conn, config.PublicConfig)
		}

		//
	})
	address := fmt.Sprintf(":%d", port)
	log.Infof("Starting websockets server on address '%s'", address)
	return http.ListenAndServe(address, nil)
}
