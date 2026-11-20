package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gruul Cluestone — Artifact {3}:
//
//	"{T}: Add {R} or {G}.
//	 {R}{G}, {T}, Sacrifice this artifact: Draw a card."
//
// Azorius Locket's shape: a two-colour mana ability, and an activated
// ability whose cost is {R}{G}, the tap and the sacrifice, all paid at
// announcement, so the Cluestone is gone before the draw resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "80911151-1e72-4d10-b00a-9fcff2dd131a",
		Name:         "Gruul Cluestone",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R|G}",
			Label:    "Add {R} or {G}",
		}},
		Activated: []ActivatedAbility{{
			Label:   "{R}{G}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{R}{G}"},
			Effect:  Do(DrawCards{N: 1}),
		}},
	})
}
