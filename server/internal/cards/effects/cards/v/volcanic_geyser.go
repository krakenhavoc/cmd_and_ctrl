package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Volcanic Geyser — Instant {X}{R}{R} (Core Set 2021, #5511):
//
//	"Volcanic Geyser deals X damage to any target."
//
// X is the value announced on the cast and charged with it; a target that is
// gone by resolution fizzles the spell (CR 608.2b), and X of 0 deals no
// damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "846a4f9c-d955-403f-8a08-5c7c3d32e180",
		Name:         "Volcanic Geyser",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetAny(),
		Purpose:      effects.ForTargets(effects.DamageXToTarget(0)),
		OnResolve: effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || ctx.X() <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: ctx.X()}.Apply(ctx)
		})),
	})
}
