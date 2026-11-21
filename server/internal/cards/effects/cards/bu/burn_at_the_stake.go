package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Burn at the Stake — Sorcery {2}{R}{R}{R} (#3961):
//
//	"As an additional cost to cast this spell, tap any number of untapped
//	 creatures you control.
//	 Burn at the Stake deals damage to any target equal to three times
//	 the number of creatures tapped this way."
//
// The additional cost is TapAnyNumberUntappedCost: the caster names any
// number of their untapped creatures on cast_spell as teamwork_ids, none
// included, and they are tapped with the spell on the stack (CR 601.2h).
// A creature that entered this turn may pay; tapping one for a cost is
// not the {T} symbol (CR 302.6).
//
// The number is read at resolution off the spell's payment record
// (ctx.TappedToPay), which holds every creature the cost tapped whatever
// has happened to them since: a creature that died in response was still
// tapped this way. A copy of the spell keeps the original's payment
// (CR 707.10), so it deals the same damage.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "eb23bbb0-dd12-44f1-8645-d92f46cbe8ff",
		Name:           "Burn at the Stake",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.TapAnyNumberUntappedCost("any number of untapped creatures you control", effects.Creature()),
		Targets:        effects.TargetAny(),
		Purpose:        effects.ForTargets(effects.DamagePerTappedToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			n := 3 * len(ctx.TappedToPay())
			if n == 0 {
				return nil
			}
			return effects.DamageToFirstTargetStep(n)(item, ctx)
		},
	})
}
