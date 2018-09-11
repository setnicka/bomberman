package main

import (
	"fmt"
	"github.com/setnicka/bomberman/board"
	"github.com/setnicka/bomberman/cell"
	"github.com/setnicka/bomberman/game"
	"github.com/setnicka/bomberman/objects"
	"github.com/setnicka/bomberman/player"
)

// Bombs!
func placeBomb(board board.Board, game *game.Game, placerState *player.State) {
	placer := game.Players[placerState]
	log.Debugf("[%s] Attempting to place bomb (%d/%d).",
		placer.Name(), placerState.Bombs, placerState.MaxBomb)

	switch {
	case placerState.Bombs > placerState.MaxBomb:
		log.Panicf("'%s' has more than it could have bombs %d/%d!", placer.Name(), placerState.Bombs, placerState.MaxBomb)
	case placerState.Bombs == 0:
		log.Debugf("'%s' cannot place bomb because player have 0 bombs.", placer.Name())
		return
	}

	placerState.Bombs--
	x, y := placerState.X, placerState.Y
	// radius is snapshot'd at this point in time
	radius := placerState.Radius

	replenishBomb := func(turn int) error {
		if placerState.Bombs < placerState.MaxBomb {
			placerState.Bombs++
		} else {
			log.Errorf("[%s] Too many bombs, %d (max %d)", placer.Name(), placerState.Bombs, placerState.MaxBomb)
		}
		return nil
	}

	doFlameout := func(turn int) error {
		log.Debugf("[%s] Bomb flameout.", placer.Name())
		removeFlame(board, x, y, radius)
		return nil
	}

	doExplosion := func(turn int) error {
		log.Debugf("[%s] Bomb exploding.", placer.Name())

		explode(game, board, x, y, radius, placerState)

		log.Debugf("[%s] Registering flameout.", placer.Name())
		game.Schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.doFlameout", placer.Name()),
			duration: 1,
			doTurn:   doFlameout,
		}, config.TurnsToFlamout)

		log.Debugf("[%s] Registering bomb replenishment.", placer.Name())
		game.Schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.replenishBomb", placer.Name()),
			duration: 1,
			doTurn:   replenishBomb,
		}, config.TurnsToReplenishUsedBomb)

		return nil
	}

	doPlaceBomb := func(turn int) error {
		board[x][y].Push(objects.Bomb)

		log.Debugf("[%s] Registering bomb explosion.", placer.Name())
		game.Schedule.Register(&BomberAction{
			name:     fmt.Sprintf("%s.doExplosion", placer.Name()),
			duration: 1,
			doTurn:   doExplosion,
		}, config.TurnsToExplode)
		return nil
	}

	game.Schedule.Register(&BomberAction{
		name:     fmt.Sprintf("%s.placeBomb", placer.Name()),
		duration: 1,
		doTurn:   doPlaceBomb,
	}, 1)

}

func explode(game *game.Game, board board.Board, explodeX, explodeY, radius int, placerState *player.State) {
	board[explodeX][explodeY].Remove(objects.Bomb)
	board.AsCross(explodeX, explodeY, radius, func(c *cell.Cell) bool {

		for playerState, player := range game.Players {
			x, y := playerState.X, playerState.Y
			if c.X == x && c.Y == y {
				log.Infof("[%s] Dying in explosion.", player.Name())
				playerState.Alive = false

				// Count points:
				placer := game.Players[placerState]
				if playerState == placerState {
					log.Infof("[%s] Receiving %d points for suicide", placer.Name(), config.PointsPerSuicide)
					placerState.Points += config.PointsPerSuicide
				} else {
					log.Infof("[%s] Receiving %d points for killing '%s'", placer.Name(), config.PointsPerKill, player.Name)
					placerState.Points += config.PointsPerKill
				}
			}
		}

		switch c.Top() {
		case objects.Wall:
			return false
		case objects.Rock:
			c.Push(objects.Flame)
			// Add points to the placer
			placerState.Points += config.PointsPerWall
			return false
		case objects.BombPU, objects.RadiusPU:
			// Explosions kill PowerUps and continues
			c.Pop()
			c.Push(objects.Flame)
			//return false
		}

		// Default action - put flame and continue
		c.Push(objects.Flame)
		return true
	})
}

func removeFlame(board board.Board, x, y, radius int) {
	board.AsCross(x, y, radius, func(c *cell.Cell) bool {
		if c.Top() == objects.Flame {
			// Remove flame
			c.Pop()
			// And remove rock if there was some
			if c.Top() == objects.Rock {
				c.Pop()
				return false
			}
			return true
		}
		return false
	})
}
