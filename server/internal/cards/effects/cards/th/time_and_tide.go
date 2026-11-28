package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Time and Tide — Instant {U}{U} (Visions, S82 older sets batch 792, #5631):
//
//	"Simultaneously, all phased-out creatures phase in and all creatures with
//	 phasing phase out."
//
// Both sets are fixed as the spell resolves, before either happens: the
// creatures that are phased out (anyone's, with what phased out indirectly
// alongside them, CR 702.26g) phase in, and the creatures on the battlefield with
// phasing phase out. A phasing creature that is phased out now phases in and
// stays in; it is not also phased out.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6fee94a-30b8-415e-8567-ad03ccb1d658",
		Name:         "Time and Tide",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.Old792TimeAndTide(ctx.Game, item)
		},
	})
}
