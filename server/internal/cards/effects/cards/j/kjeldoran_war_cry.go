package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kjeldoran War Cry — Instant {1}{W} (Coldsnap, #5597):
//
//	"Creatures you control get +X/+X until end of turn, where X is 1 plus the
//	 number of cards named Kjeldoran War Cry in all graveyards."
//
// X is counted once as the spell resolves (CR 608.2h): this War Cry is still on
// the stack, so it is not counted, and every player's graveyard is. The creatures
// that get it are the ones its controller controls then (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e45a1626-d659-4eda-bc6b-bd8119ce7386",
		Name:         "Kjeldoran War Cry",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			x := 1
			for _, p := range ctx.Game.Seats {
				if p == nil || p.Graveyard == nil {
					continue
				}
				for _, c := range p.Graveyard.Cards {
					if c.Name == "Kjeldoran War Cry" {
						x++
					}
				}
			}
			return effects.BoostUntilEOT{Match: effects.And(effects.Creature(), effects.YouControl()), Power: x, Toughness: x, Label: "Kjeldoran War Cry — +X/+X"}.Apply(ctx)
		},
	})
}
