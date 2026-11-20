package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Golgari Cluestone — Artifact {3}:
//
//	"{T}: Add {B} or {G}.
//	 {B}{G}, {T}, Sacrifice this artifact: Draw a card."
//
// Azorius Locket's shape: a two-colour mana ability, and an activated
// ability whose cost is {B}{G}, the tap and the sacrifice, all paid at
// announcement, so the Cluestone is gone before the draw resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e917ad3d-df8e-442b-918c-3aa7a10f3cc4",
		Name:         "Golgari Cluestone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{B|G}",
			Label:    "Add {B} or {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{B}{G}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{B}{G}"},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
