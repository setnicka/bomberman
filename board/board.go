package board

import (
	"github.com/nsf/termbox-go"
	"github.com/setnicka/bomberman/cell"
	"github.com/setnicka/bomberman/game"
	"github.com/setnicka/bomberman/objects"
	"github.com/setnicka/bomberman/player"
	"math/rand"
)

type Board [][]*cell.Cell

func newBoard(x, y int) Board {
	b := make(Board, x)
	for i := range b {
		b[i] = make([]*cell.Cell, y)
	}
	return b
}

func SetupBoard(g *game.Game, x, y, rockFreeRadius int, rockDensity float64) Board {
	board := newBoard(x, y)

	freeCells := board.putwWalls()
	rockPlaced := board.setupRocks(freeCells, rockDensity)
	cleared := board.clearAroundPlayers(g.Players, rockFreeRadius)
	rockPlaced -= cleared

	bombRocksLeft := rockPlaced / 2
	radiusRocksLeft := rockPlaced / 2

	onlyRocks := func(c *cell.Cell) bool { return c.Top() == objects.Rock }

	putPwrUpUnder := func(c *cell.Cell) {
		rock, _ := c.Pop()
		switch rand.Intn(2) {
		case 0:
			g.TryPutRadiusPU(c, radiusRocksLeft)
			radiusRocksLeft--
		case 1:
			g.TryPutBombPU(c, bombRocksLeft)
			bombRocksLeft--
		}
		c.Push(rock)
	}
	board.filter(onlyRocks, putPwrUpUnder)

	return board
}

func (b Board) putwWalls() (free int) {
	b.forEachIndex(func(_ *cell.Cell, x, y int) {
		b[x][y] = cell.NewCell(objects.Ground, x, y)
		switch {
		case
			x == 0 || x == len(b)-1,    // Left and right
			y == 0 || y == len(b[0])-1, // Top and bottom
			y%2 == 0 && x%2 == 0:       // Every second cell
			b[x][y].Push(objects.Wall)
		default:
			free++
		}
	})
	return
}

func (b Board) setupRocks(freeCells int, densityPercent float64) int {
	needRock := int(float64(freeCells) * densityPercent)
	rockPlaced := 0
	rockProb := func(rockLeft, freeCell int) float64 {
		return float64(rockLeft) / float64(freeCell)
	}

	groundTest := func(c *cell.Cell) bool { return c.Top() == objects.Ground }

	b.filter(groundTest, func(c *cell.Cell) {
		prob := rockProb(needRock, freeCells)
		roll := rand.Float64()
		if roll < prob {
			needRock--
			rockPlaced++
			c.Push(objects.Rock)
		}
		freeCells--
	})
	return rockPlaced
}

func (b Board) clearAroundPlayers(players map[*player.State]player.Player, radius int) (removed int) {
	for state := range players {
		if !state.Alive {
			continue
		}

		x, y := state.X, state.Y
		b.AsSquare(x, y, radius, func(cell *cell.Cell) {
			if cell.Top() == objects.Rock {
				cell.Pop()
				removed++
			}
		})
		b[x][y].Push(state.GameObject)
	}
	return
}

func (b Board) Traversable(x, y int) bool {
	if x < 0 || y < 0 {
		return false
	}
	return b[x][y].Top().Traversable()
}

func (b Board) Draw(players map[*player.State]player.Player, shift ...int) {
	shiftY := 0
	if len(shift) > 0 {
		shiftY = shift[0]
	}

	b.forEach(func(c *cell.Cell) {
		c.Top().Draw(c.X, c.Y+shiftY)
	})

	for state := range players {
		if !state.Alive {
			b[state.X][state.Y].Remove(state.GameObject)
			continue
		}
	}

	termbox.Flush()
}

func (b Board) Export() player.ExportedBoard {
	exported := make([]string, len(b))
	for i := range exported {
		// Create slice of runes
		row := []rune{}
		for _, cell := range b[i] {
			row = append(row, cell.Top().Symbol())
		}
		exported[i] = string(row)
	}
	return player.ExportedBoard(exported)
}

///////////
// Helpers

// functional iterations

func (b Board) forEachIndex(apply func(*cell.Cell, int, int)) {
	for x, inner := range b {
		for y, cell := range inner {
			apply(cell, x, y)
		}
	}
}

func (b Board) forEach(apply func(*cell.Cell)) {
	b.forEachIndex(func(c *cell.Cell, x, y int) { apply(c) })
}

func (b Board) filter(test func(*cell.Cell) bool, apply func(*cell.Cell)) {
	b.forEach(func(cell *cell.Cell) {
		if test(cell) {
			apply(cell)
		}
	})
}

func (b Board) AsSquare(x, y, rad int, apply func(*cell.Cell)) {
	for i := max(x-rad, 0); i <= min(x+rad, len(b)-1); i++ {
		for j := max(y-rad, 0); j <= min(y+rad, len(b[0])-1); j++ {
			apply(b[i][j])
		}
	}
}

func (b Board) AsCross(x, y, dist int, apply func(*cell.Cell) bool) {
	// (x,y) and to the right
	var c *cell.Cell
	for i := x; i < min(x+dist, len(b)); i++ {
		c = b[i][y]
		if !apply(c) {
			break
		}
	}

	// left of (x,y)
	for i := x - 1; i > max(x-dist, 0); i-- {
		c = b[i][y]
		if !apply(c) {
			break
		}
	}

	// below (x,y)
	for j := y + 1; j < min(y+dist, len(b)); j++ {
		c = b[x][j]
		if !apply(c) {
			break
		}
	}

	// above (x,y)
	for j := y - 1; j > max(y-dist, 0); j-- {
		c = b[x][j]
		if !apply(c) {
			break
		}
	}
}

// Integer math

func min(n, m int) int {
	if n < m {
		return n
	}
	return m
}

func max(n, m int) int {
	if n > m {
		return n
	}
	return m
}
