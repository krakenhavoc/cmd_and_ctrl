package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tempered Veteran — Creature — Human Knight {1}{W}, 1/2 (Core Set 2021, #5511):
//
//	"{W}, {T}: Put a +1/+1 counter on target creature with a +1/+1 counter
//	 on it.
//	 {4}{W}{W}, {T}: Put a +1/+1 counter on target creature."
//
// Dragon Blood's counter (PlusOneCounterOnTheTarget) twice: the cheap row
// takes Helium Squirter's "with a +1/+1 counter on it" clause, judged at
// announce and again as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c19f3f00-2c7b-4e85-a034-e65760d9cc13",
		Name:         "Tempered Veteran",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{W}, {T}: Put a +1/+1 counter on target creature with a +1/+1 counter on it.",
				Cost:    effects.Plus(effects.ManaCost("{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature with a +1/+1 counter on it", effects.Not(effects.B13WithoutPlusOneCounter())),
				Effect:  effects.PlusOneCounterOnTheTarget,
			},
			{
				Label:   "{4}{W}{W}, {T}: Put a +1/+1 counter on target creature.",
				Cost:    effects.Plus(effects.ManaCost("{4}{W}{W}"), effects.TapCost()),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.PlusOneCounterOnTheTarget,
			},
		},
	})
}
