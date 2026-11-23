package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sea Gate Wreckage — Land (EDHREC rank 5148):
//
//	"{T}: Add {C}. ({C} represents colorless mana.)
//	 {2}{C}, {T}: Draw a card. Activate only if you have no cards in
//	 hand."
//
// A hellbent draw land. The {C} in the cost is colorless mana
// specifically, which the cost parser already distinguishes from
// generic, so a Wastes or a Sol Ring pays it and an Island does not.
// "Activate only if you have no cards in hand" is the activation
// condition (CR 602.1b, #743), NoCardsInHand — the hand's size, which
// is public.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91f34686-cb96-49c0-b4a7-49dd1fd076e2",
		Name:         "Sea Gate Wreckage",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}{C}, {T}: Draw a card. Activate only if you have no cards in hand.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{2}{C}"), effects.TapCost()),
			Condition: effects.NoCardsInHand(),
			Effect:    effects.B36DrawOne,
		}},
	})
}
