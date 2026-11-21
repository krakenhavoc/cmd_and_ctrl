package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "d76fc3aa-3962-4d64-a082-765e09bcdf25",
		Name:         "Kami's Flare",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			victim, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			controller := victim.Controller
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if !effects.NeoB10ControlsModified(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: 2}.Apply(ctx)
		},
	})
}
