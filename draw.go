package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"

	"github.com/setnicka/bomberman/board"
	"github.com/setnicka/bomberman/player"
)

const (
	PLAYER_X_OFFSET = 25
	PLAYER_Y_OFFSET = 8
	PLAYERS_IN_LINE = 6
	FG              = termbox.ColorWhite
	BG              = termbox.ColorBlack
)

func consoleDraw(board board.Board, players map[*player.State]player.Player, drawMap bool) {
	// Draw statistics

	// 1. Get states
	states := []*player.State{}
	for state := range players {
		if !state.Hidden {
			states = append(states, state)
		}
	}

	// 2. Sort them
	sort.Slice(states, func(i, j int) bool {
		return (strings.Compare(states[i].Name, states[j].Name) == -1)
	})

	// 3. Draw players
	line := 0
	in_line := 0
	for _, state := range states {
		if in_line == PLAYERS_IN_LINE {
			line++
			in_line = 0
		}
		drawPlayer(state, in_line*PLAYER_X_OFFSET, line*PLAYER_Y_OFFSET)
		in_line++
	}

	if drawMap {
		board.Draw(players, (line+1)*PLAYER_Y_OFFSET+2)
	} else {
		termbox.Flush()
	}
}

func tbprint(x, y int, fg, bg termbox.Attribute, msg string) {
	for _, c := range msg {
		termbox.SetCell(x, y, c, fg, bg)
		x += runewidth.RuneWidth(c)
	}
}

func drawPlayer(state *player.State, x int, y int) {
	name := fmt.Sprintf("[%s]%s", state.Symbol, state.Name)
	if state.Alive {
		tbprint(x, y, termbox.ColorGreen, BG, name)
	} else {
		tbprint(x, y, termbox.ColorRed, BG, name)
	}
	tbprint(x, y+1, FG, BG, fmt.Sprintf("      Type: %s   ", state.Type))
	tbprint(x, y+2, FG, BG, fmt.Sprintf("     Bombs: %d/%d", state.Bombs, state.MaxBomb))
	tbprint(x, y+3, FG, BG, fmt.Sprintf("    Radius: %d   ", state.Radius))
	tbprint(x, y+4, FG, BG, fmt.Sprintf("    Points: %d   ", state.Points))
	tbprint(x, y+5, FG, BG, fmt.Sprintf("  Total p.: %d   ", state.TotalPoints))
	if state.Type == "websocket" {
		if state.Connected {
			tbprint(x, y+6, termbox.ColorGreen, BG, fmt.Sprintf("Resp. time: %v   ", state.ResponseTime))
		} else {
			tbprint(x, y+6, termbox.ColorRed, BG, "Not connected")
		}
	}
}
