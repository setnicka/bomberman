package main

import (
	"github.com/gorilla/websocket"
	"time"

	"github.com/setnicka/bomberman/player"
)

type Config struct {
	PublicConfig

	Width       int     `json:"width"`
	Height      int     `json:"height"`
	RockDensity float64 `json:"rock_density"`

	TurnDuration int `json:"turn_duration_ms"`

	FreeAreaAroundPlayers int `json:"free_area_around_players"`
	DefaultMaxBombs       int `json:"default_max_bombs"`
	DefaultBombRadius     int `json:"default_bomb_radius"`
	TotalRadiusPowerups   int `json:"total_radius_powerups"`
	TotalBombsPowerups    int `json:"total_bombs_powerups"`
}

type PublicConfig struct {
	TurnsToFlamout           int `json:"turns_to_flamout"`
	TurnsToReplenishUsedBomb int `json:"turns_to_replenish_used_bomb"`
	TurnsToExplode           int `json:"turns_to_explode"`

	PointsPerWall    int `json:"points_per_wall"`
	PointsPerKill    int `json:"points_per_kill"`
	PointsPerSuicide int `json:"points_per_suicide"`
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
