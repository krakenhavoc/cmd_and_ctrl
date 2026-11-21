package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Selesnya Cluestone — Artifact {3}:
//
//	"{T}: Add {G} or {W}.
//	 {G}{W}, {T}, Sacrifice this artifact: Draw a card."
//
// Azorius Locket's shape: a two-colour mana ability, and an activated
// ability whose cost is {G}{W}, the tap and the sacrifice, all paid at
// announcement, so the Cluestone is gone before the draw resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "792d4afd-b3de-4be2-b68a-03a5be7b9627",
		Name:         "Selesnya Cluestone",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:     effects.ManaAbilityCost{Tap: true},
			Produced: "{G|W}",
			Label:    "Add {G} or {W}",
		}},
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}{W}, {T}, Sacrifice this artifact: Draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    game.AbilityCost{Tap: true, SacrificeSelf: true, Mana: "{G}{W}"},
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
