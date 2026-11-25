package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Memorial to Unity — Land (Dominaria, S82 OLD batch 215, #5530):
//
//	"This land enters tapped.
//	 {T}: Add {G}.
//	 {2}{G}, {T}, Sacrifice this land: Look at the top five cards of your
//	 library. You may reveal a creature card from among them and put it into
//	 your hand. Then put the rest on the bottom of your library in a random
//	 order."
//
// The green Memorial (Memorial to Genius's shape): a tapped Forest that cashes
// in for Brightwood Tracker's look-and-take over five cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a74494ef-aa35-4830-9b4c-47bff5270efc",
		Name:         "Memorial to Unity",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label: "{2}{G}, {T}, Sacrifice this land: Look at the top five cards of your library. You may reveal a creature card " +
				"from among them and put it into your hand. Then put the rest on the bottom of your library in a random order.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost(), effects.SacrificeThis()),
			Effect: effects.LookAtTopThenMayTakeToHand(5, effects.Creature(), 1,
				"Memorial to Unity — you may reveal a creature card and put it into your hand"),
		}},
	})
}
