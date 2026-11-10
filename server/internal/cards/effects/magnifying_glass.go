package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Magnifying Glass — Artifact {3} (Murders at Karlov Manor, #4469):
//
//	"{T}: Add {C}.
//	 {4}, {T}: Investigate. (Create a Clue token. It's an artifact with
//	 "{2}, Sacrifice this token: Draw a card.")"
//
// A colorless mana rock with a second, stack-using ability: investigate
// is "create a Clue token" (CR 701.16a), the catalog's Clue.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2ce9e55-07c0-4f36-a59e-127693fc0f62",
		Name:         "Magnifying Glass",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}: Investigate.",
			Purpose: game.Purpose{Answers: game.AnswerValue, Tokens: 1},
			Cost:    Plus(ManaCost("{4}"), TapCost()),
			Effect:  Do(Investigate{N: 1}),
		}},
	})
}
