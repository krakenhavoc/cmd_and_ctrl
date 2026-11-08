package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// tdm_df_helpers.go — shared reads for Tarkir: Dragonstorm cards from D to
// F (#4463).

// tdmControlsCreatureWithACounter is Tarkir: Dragonstorm's "if you control
// a creature with a counter on it" (Delta Bloodflies, Formation Breaker):
// any kind of counter, on a creature the player controls, read off the
// battlefield as it stands.
func tdmControlsCreatureWithACounter(g *game.Game, player uuid.UUID) bool {
	for i := range g.Battlefield.Cards {
		c := &g.Battlefield.Cards[i]
		if c.Controller != player || !c.IsCreature() {
			continue
		}
		for _, n := range c.Counters {
			if n > 0 {
				return true
			}
		}
	}
	return false
}
