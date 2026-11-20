package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "d001321d-b272-4f94-a333-7cfe8482da20",
		Name:         "Decommission",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := destroyTheTargetPermanent(item, ctx); err != nil {
				return err
			}
			if !Revolt(ctx.Game, ctx.Controller()) {
				return nil
			}
			return GainLife{Player: ctx.Controller(), Amount: 3}.Apply(ctx)
		},
	})
}
