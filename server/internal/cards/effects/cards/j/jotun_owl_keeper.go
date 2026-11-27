package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jötun Owl Keeper — Creature — Giant {2}{W}, 3/3 (Coldsnap, S82 OLD b558):
//
//	"Cumulative upkeep {W} or {U}
//	 When this creature dies, create a 1/1 white Bird creature token with
//	 flying for each age counter on it."
//
// "{W} or {U}" is one hybrid symbol per age counter (CR 702.24a: each choice
// is made separately for each age counter, #6189). The Birds are counted from
// the Keeper's last-known age counters (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3acef3dc-61ee-48e0-9b2e-a8e29d5f4ac7",
		Name:         "Jötun Owl Keeper",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.CumulativeUpkeep("Jötun Owl Keeper — cumulative upkeep {W} or {U}", "{W/U}"),
			effects.WhenThisDies("Jötun Owl Keeper — create a 1/1 white Bird creature token with flying for each age counter on it",
				func(g *game.Game, item *game.StackItem) error {
					n := effects.AgeCountersAsItLeft(g, item)
					if n <= 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 white Bird with flying"), N: n}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
