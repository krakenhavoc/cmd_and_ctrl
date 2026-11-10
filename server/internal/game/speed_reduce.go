package game

import "github.com/google/uuid"

// speed_reduce.go — "reduce that player's speed by 1. This effect can't
// reduce their speed below 1" (Spikeshell Harrier, Aetherdrift).
//
// CR 702.179 says nothing ever lowers a speed except an effect that says
// so, and this is that effect's verb. setSpeedLocked stays the one write
// (an EventSpeedChanged on every real change; a layer input, so a "Max
// speed —" static follows a speed that falls from 4).

// SpeedGreaterThanEachOtherForEffect reports whether `player`'s speed is
// greater than the speed of each other player still in the game ("if
// that opponent's speed is greater than each other player's speed"). A
// player with no speed has speed 0, so a table where nobody else has
// speed answers true for anyone who has some, and never for a player with
// none. Caller must hold g.mu.
func (g *Game) SpeedGreaterThanEachOtherForEffect(player uuid.UUID) bool {
	mine := g.SpeedOf(player)
	if mine <= 0 {
		return false
	}
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == player {
			continue
		}
		if p.Speed >= mine {
			return false
		}
	}
	return true
}

// ReduceSpeedForEffect lowers `player`'s speed by `by`, never below
// `floor`: a speed already at or under the floor stays as it is (the
// effect never raises it). Returns the speed afterwards. Caller must hold
// g.mu in write mode.
func (g *Game) ReduceSpeedForEffect(player uuid.UUID, by, floor int) int {
	cur := g.SpeedOf(player)
	if by <= 0 || cur <= floor {
		return cur
	}
	next := cur - by
	if next < floor {
		next = floor
	}
	g.setSpeedLocked(player, next)
	return g.SpeedOf(player)
}
