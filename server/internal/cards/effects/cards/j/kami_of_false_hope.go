package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kami of False Hope — Creature — Spirit {W}, 1/1 (Betrayers of Kamigawa,
// #5602):
//
//	"Sacrifice this creature: Prevent all combat damage that would be
//	 dealt this turn."
//
// Fog on a body: the sacrifice is the cost, and the effect is Fog's
// PreventAllCombatDamageThisTurn, a prevention shield over every combat
// damage event for the rest of the turn (CR 615.1). Damage that is not
// combat damage is untouched.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "49984248-4800-4f46-933a-fedd970ae910",
		Name:         "Kami of False Hope",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Prevent all combat damage that would be dealt this turn.",
			Purpose: game.Purpose{Answers: game.AnswerPrevent | game.AnswerSacOutlet},
			Cost:    effects.SacrificeThis(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.PreventAllCombatDamageThisTurn{Label: "Kami of False Hope: prevent combat damage"}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
