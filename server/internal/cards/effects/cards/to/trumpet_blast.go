package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Trumpet Blast — Instant {2}{R} (Urza's Destiny, #5621):
//
//	"Attacking creatures get +2/+0 until end of turn."
//
// Army of Allah at {2}{R}. The attacking creatures are read as the
// spell resolves: every one, whoever controls it, gets +2/+0, and a
// creature that starts attacking later in the turn does not. Outside
// combat nothing is attacking, so it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3b408e7-4d39-47a3-b00c-846903f8ae77",
		Name:         "Trumpet Blast",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: 2, Label: "Trumpet Blast — attacking creatures get +2/+0"}.Apply(ctx)
		},
	})
}
