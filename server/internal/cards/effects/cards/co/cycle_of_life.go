package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cycle of Life — Enchantment {1}{G}{G} (Mirage):
//
//	"Return this enchantment to its owner's hand: Target creature you cast this
//	 turn has base power and toughness 0/1 until your next upkeep. At the
//	 beginning of your next upkeep, put a +1/+1 counter on that creature."
//
// The enchantment returns to hand as a cost (CR 602.2b). The target is a creature
// permanent that a spell the controller cast this turn became (its provenance and
// this turn's cast events, OldB804CreatureYouCastThisTurn). The base 0/1 lasts
// until the controller's next upkeep begins; the delayed trigger names the
// creature as the object it was (CR 400.7, 603.7c) and puts the counter on it
// only if it is still that permanent.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5941374b-eeb8-4d0e-9367-f7f0f8b79739",
		Name:         "Cycle of Life",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Return this enchantment to its owner's hand: Target creature you cast this turn has base power and toughness 0/1 until your next upkeep. At the beginning of your next upkeep, put a +1/+1 counter on that creature.",
			Cost:    effects.ReturnThis(),
			Targets: effects.TargetCreature("target creature you cast this turn", effects.OldB804CreatureYouCastThisTurn()),
			Effect:  effects.OldB804CycleOfLifeResolve,
		}},
	})
}
