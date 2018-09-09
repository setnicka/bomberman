package main

import (
	"github.com/gorilla/websocket"
	"time"

	"github.com/setnicka/bomberman/player"
)

type Config struct {
	Width       int     `json:"width"`
	Height      int     `json:"height"`
	RockDensity float64 `json:"rock_density"`

	TurnDuration             int `json:"turn_duration_ms"`
	TurnsToFlamout           int `json:"turns_to_flamout"`
	TurnsToReplenishUsedBomb int `json:"turns_to_replenish_used_bomb"`
	TurnsToExplode           int `json:"turns_to_explode"`

	FreeAreaAroundPlayers int `json:"free_area_around_players"`
	DefaultMaxBombs       int `json:"default_max_bombs"`
	DefaultBombRadius     int `json:"default_bomb_radius"`
	TotalRadiusPowerups   int `json:"total_radius_powerups"`
	TotalBombsPowerups    int `json:"total_bombs_powerups"`
}

type PlayersConf []PlayerConf

type PlayerConf struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Symbol   string `json:"symbol"`
	Password string `json:"password"`
	StartX   int    `json:"startX"`
	StartY   int    `json:"startY"`
}

type RemotePlayer struct {
	playerConf PlayerConf
	state      player.State

	// Outer communication
	updateChan  chan player.State
	moveChan    chan player.Move
	outMoveChan chan player.Move

	// State distribution to clients
	clientChannels       map[*RemotePlayerClient]chan *player.State
	clientUnregisterChan chan *RemotePlayerClient
	lastClientId         int

	responseTime time.Duration
}

type RemotePlayerClient struct {
	Id           int
	RemotePlayer *RemotePlayer
	Conn         *websocket.Conn
	updateChan   chan *player.State
	moveChan     chan player.Move
}
