package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coeurl — Creature — Cat Beast {1}{W}, 2/2 (Final Fantasy, #4462):
//
//	"{1}{W}, {T}: Tap target nonenchantment creature."
//
// Arashin Sunshield's tapper at {1}{W}, with the target narrowed to a
// creature that is not also an enchantment, read on the post-layer type
// line (an enchantment creature such as a Theros god or a Saga creature
// can't be chosen). The {T} in the cost waits out summoning sickness
// (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00d1596a-c3e2-4109-86da-388934a0c652",
		Name:         "Coeurl",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}, {T}: Tap target nonenchantment creature.",
			Cost:    effects.Plus(effects.ManaCost("{1}{W}"), effects.TapCost()),
			Targets: effects.TargetCreature("target nonenchantment creature", effects.Not(effects.Enchantment())),
			Effect:  effects.TapChosenPermanent,
		}},
	})
}
