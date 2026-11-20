package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scarecrow Guide — Artifact Creature — Scarecrow {2}, 2/1 (Wilds of
// Eldraine, #4471):
//
//	"Reach
//	 {1}: Add one mana of any color. Activate only once each turn."
//
// The filter is Barrels of Blasting Jelly's: a mana ability with the
// engine's OncePerTurn bit, so every path refuses a second activation in
// a turn.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "cf7bb5d4-c33f-4879-a793-6d6f2a4d94be",
		Name:            "Scarecrow Guide",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		ManaAbilities: []effects.ManaAbility{{
			Cost:        effects.ManaAbilityCost{Mana: "{1}"},
			Produced:    "{W|U|B|R|G}",
			Label:       "{1}: Add one mana of any color. Activate only once each turn.",
			OncePerTurn: true,
			Answers:     game.AnswerValue,
		}},
	})
}
