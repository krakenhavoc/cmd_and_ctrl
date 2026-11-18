package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Kami's Flare — Instant {1}{R} (Kamigawa: Neon Dynasty):
//
//	"Kami's Flare deals 3 damage to target creature or planeswalker. Kami's
//	 Flare also deals 2 damage to that permanent's controller if you control
//	 a modified creature. (Equipment, Auras you control, and counters are
//	 modifications.)"
//
// Read as the spell resolves (no "as you cast"). The permanent's controller
// is taken before the damage, so one that dies still has its controller hit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d76fc3aa-3962-4d64-a082-765e09bcdf25",
		Name:         "Kami's Flare",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(DamageToTarget(0, 3)),
		Targets:      TargetPermanent("target creature or planeswalker", Or(Creature(), Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := victim.Controller
			if err := (DealDamage{Source: ctx.Source(), Target: id, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if !neoB10ControlsModified(ctx.Game, ctx.Controller()) {
				return nil
			}
			return DealDamage{Source: ctx.Source(), Target: controller, Amount: 2}.Apply(ctx)
		},
	})
}
