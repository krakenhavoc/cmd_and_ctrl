package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "e917ad3d-df8e-442b-918c-3aa7a10f3cc4",
		Name:         "Golgari Cluestone",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{B|G}",
			Label:    "Add {B} or {G}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{B}{G}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{B}{G}"},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
