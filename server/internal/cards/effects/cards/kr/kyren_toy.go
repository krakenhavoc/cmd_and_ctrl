package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kyren Toy — Artifact {3} (Mercadian Masques, S82 OLD batch 698, #5619):
//
//	"{1}, {T}: Put a charge counter on this artifact.
//	 {T}, Remove X charge counters from this artifact: Add an amount of {C}
//	 equal to X plus one."
//
// Black Mana Battery's shape: the banking is an activated ability (it adds no
// mana, so it uses the stack), and the payout is a mana ability whose cost
// removes X charge counters, zero included, announced and read back through the
// paid-cost record. It adds {C} for each counter removed, plus one.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f1a0d112-9e1b-4354-bd7f-0051bff8f904",
		Name:         "Kyren Toy",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				RemoveCounters: effects.RemoveCountersXFromThis(game.CounterCharge, 0).RemoveCounters,
			},
			ProducedForPaid: effects.S82OldB698ColorlessForXPlusOne,
			Label:           "{T}, Remove X charge counters: Add an amount of {C} equal to X plus one",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}: Put a charge counter on this artifact.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost()),
			Effect:  effects.PutChargeCounterOnThis,
		}},
	})
}
