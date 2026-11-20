package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Simic Cluestone — Artifact {3}:
//
//	"{T}: Add {G} or {U}.
//	 {G}{U}, {T}, Sacrifice this artifact: Draw a card."
//
// Azorius Locket's shape: a two-colour mana ability, and an activated
// ability whose cost is {G}{U}, the tap and the sacrifice, all paid at
// announcement, so the Cluestone is gone before the draw resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b65050d-a9cd-4b6c-9114-af89bfa6682f",
		Name:         "Simic Cluestone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G|U}",
			Label:    "Add {G} or {U}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{G}{U}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{G}{U}"},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
