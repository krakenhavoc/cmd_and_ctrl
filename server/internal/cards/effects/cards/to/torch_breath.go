package cardsto

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Torch Breath — Instant {X}{R} (Streets of New Capenna, #5491):
//
//	"This spell costs {2} less to cast if it targets a blue permanent.
//	 This spell can't be countered.
//	 Torch Breath deals X damage to target creature or planeswalker."
//
// The reduction reads the announced target (CostsLessIfItTargets) and
// comes off the generic part of {X}{R}, after X is chosen (CR 601.2f).
// X is the announced X; the damage is the spell's, dealt only while the
// target is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "b6d61d96-a56a-4f4c-a196-05e03fe16ed2",
		Name:            "Torch Breath",
		Completeness:    effects.CompletenessFull,
		XMatters:        true,
		CantBeCountered: true,
		Targets:         effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		Purpose:         effects.ForTargets(effects.DamageXToTarget(0)),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(2, "This spell costs {2} less to cast if it targets a blue permanent.",
				effects.OfColor("U")),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil || ctx.X() <= 0 {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: target, Amount: ctx.X()}.Apply(ctx)
		},
	})
}
