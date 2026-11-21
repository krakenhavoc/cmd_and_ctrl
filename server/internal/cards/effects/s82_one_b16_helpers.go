package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_one_b16_helpers.go — S82 set ONE, batch 16 (toxic and corrupted
// rares). Names start s82OneB16.

// s82OneB16CorruptedOpponents is "the number of opponents who have three
// or more poison counters": opponents
// of `controller` still in the game with at least three. A player who has
// left the game is not an opponent.
func S82OneB16CorruptedOpponents(g *game.Game, controller uuid.UUID) int {
	n := 0
	for _, p := range g.Seats {
		if p == nil || p.Eliminated || p.ID == controller {
			continue
		}
		if p.Poison >= 3 {
			n++
		}
	}
	return n
}
