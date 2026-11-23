package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Salvaged Manaworker — Artifact Creature — Construct {2}, 1/3 (Dominaria
// United):
//
//	"{1}: Add one mana of any color. Activate only once each turn."
//
// Barrels of Blasting Jelly's mana ability: the engine's OncePerTurn bit
// refuses a second activation in a turn on every path (CR 605.1a, 602.5b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a96814fe-dc74-4d02-a44e-f7daa794b647",
		Name:         "Salvaged Manaworker",
		Completeness: effects.CompletenessFull,
		ManaAbilities: []effects.ManaAbility{{
			Cost:        effects.ManaAbilityCost{Mana: "{1}"},
			Produced:    "{W|U|B|R|G}",
			Label:       "{1}: Add one mana of any color. Activate only once each turn.",
			OncePerTurn: true,
			Answers:     game.AnswerValue,
		}},
	})
}
