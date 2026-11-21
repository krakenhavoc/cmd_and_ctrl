package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Decommission — Instant {2}{W} (Aether Revolt,
// #5536):
//
//	"Destroy target artifact or enchantment.
//	 Revolt — If a permanent left the battlefield under your control this
//	 turn, you gain 3 life."
//
// Disenchant, then Fatal Push's revolt read at resolution (revolt.go).
// The permanent the spell destroys counts only if its controller is the
// caster, as printed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d001321d-b272-4f94-a333-7cfe8482da20",
		Name:         "Decommission",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target artifact or enchantment", effects.Or(effects.Artifact(), effects.Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			if !effects.Revolt(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		},
	})
}
