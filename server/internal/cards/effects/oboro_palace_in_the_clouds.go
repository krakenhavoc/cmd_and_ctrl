package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Oboro, Palace in the Clouds — Legendary Land (Saviors of Kamigawa,
// #5601):
//
//	"{T}: Add {U}.
//	 {1}: Return Oboro to its owner's hand."
//
// The return is an ordinary activated ability, not a mana ability, so it
// uses the stack; it returns the land only if it is still the same object
// (BounceToHand's CR 400.7 guard). It answers removal aimed at the land by
// saving it (AnswerSurvive).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "645fb11b-d684-4bec-8532-8fa97e8f7b28",
		Name:         "Oboro, Palace in the Clouds",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{U}",
			Label:    "Add {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{1}: Return Oboro to its owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    ManaCost("{1}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return BounceToHand{Target: item.SourceCardID}.Apply(NewContext(g, item))
			},
		}},
	})
}
