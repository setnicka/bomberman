package player

import (
	"time"

	"github.com/setnicka/bomberman/cell"
)

type ExportedBoard []string

// BasicState holds basic fields related to the one player
type BasicState struct {
	Name                   string
	Number                 int
	X, Y, LastX, LastY     int
	Bombs, MaxBomb, Radius int
	Alive                  bool
	Points                 int
	TotalPoints            int `json:"-"`
}

// State holds all info for one player (including link to exported board and to the other players).
// This is used for exporting whole player's state
type State struct {
	BasicState
	Turn         int
	Board        *ExportedBoard
	GameObject   cell.GameObject
	Players      []*BasicState
	Message      string
	Type         string        `json:"-"`
	Symbol       string        `json:"-"`
	Hidden       bool          `json:"-"`
	ResponseTime time.Duration `json:"-"`
	Connected    bool          `json:"-"`
}

type Move string

const (
	Up      = Move("up")
	Down    = Move("down")
	Left    = Move("left")
	Right   = Move("right")
	PutBomb = Move("bomb")
)

type Player interface {
	Name() string
	Move() <-chan Move
	Update() chan<- State
}
