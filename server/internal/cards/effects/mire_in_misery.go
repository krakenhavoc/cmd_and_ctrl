package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "da1d1f64-e239-4b92-9eb9-fe9ea06aa18a",
		Name:         "Mire in Misery",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return EachPlayerSacrifices{
				ExceptController: true,
				Match:            Or(Creature(), Enchantment()),
				Label:            "a creature or enchantment",
			}.Apply(ctx)
		},
	})
}
