package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scrollshift — Instant {2}{W} (March of the Machine, S82 #5477):
//
//	"Exile up to one target artifact, creature, or enchantment you control,
//	 then return it to the battlefield under its owner's control.
//	 Draw a card."
//
// "Up to one": with no target chosen the spell still resolves and you still
// draw (CR 601.2c, 608.2b). The permanent returns as a new object (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9f762b9f-7525-4ff8-bc1d-fbab66fd4013",
		Name:         "Scrollshift",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		Targets: effects.TargetPermanent("up to one target artifact, creature, or enchantment you control",
			effects.And(effects.Or(effects.Artifact(), effects.Creature(), effects.Enchantment()), effects.YouControl())).WithCount(0, 1),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.Flicker{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.DrawCards{N: 1}.Apply(ctx)
		},
	})
}
