package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Depths of Desire — Instant {2}{U} (Ixalan, S82 older sets):
//
//	"Return target creature to its owner's hand. Create a Treasure token."
//
// The bounce, then the Treasure. The Treasure is not conditional on the
// creature having been returned (a commander sent to the command zone, or a
// target that has gone, only changes the first sentence), so it is made
// whenever the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "caa0c20c-594b-4ffd-87de-60bf15d8c3be",
		Name:         "Depths of Desire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if id, ok := effects.B16FirstLegalTargetCard(ctx); ok {
				if err := (effects.BounceToHand{Target: id}).Apply(ctx); err != nil {
					return err
				}
			}
			return effects.CreateToken{Template: effects.TreasureToken(), N: 1}.Apply(ctx)
		},
	})
}
