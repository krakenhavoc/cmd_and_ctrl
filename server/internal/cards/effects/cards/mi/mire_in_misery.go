package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mire in Misery — Sorcery {1}{B} (Commander 2019, #5517):
//
//	"Each opponent sacrifices a creature or enchantment of their choice."
//
// An edict that does not target: each opponent, in APNAP order, picks one
// creature or enchantment they control (EachPlayerSacrifices with the caster
// left out). An opponent with neither sacrifices nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "da1d1f64-e239-4b92-9eb9-fe9ea06aa18a",
		Name:         "Mire in Misery",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.Or(effects.Creature(), effects.Enchantment()),
				Label:            "a creature or enchantment",
			}.Apply(ctx)
		},
	})
}
