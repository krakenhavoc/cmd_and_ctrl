package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perilous Predicament — {4}{B} Instant (Aether Revolt, #5536):
//
//	"Each opponent sacrifices an artifact creature and a nonartifact
//	 creature of their choice."
//
// Each opponent names the two creatures themselves (the two sets cannot
// overlap, so the order of the two prompts changes nothing): first every
// opponent's artifact creature, then every opponent's nonartifact one. An
// opponent short of either type sacrifices what they can.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "169728d6-7820-4a6d-a815-70714ad3e887",
		Name:         "Perilous Predicament",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.EachPlayerSacrifices{
					ExceptController: true,
					Match:            effects.And(effects.Creature(), effects.Artifact()),
					Label:            "an artifact creature",
				}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.EachPlayerSacrifices{
					ExceptController: true,
					Match:            effects.And(effects.Creature(), effects.Not(effects.Artifact())),
					Label:            "a nonartifact creature",
				}.Apply(ctx)
			}),
		),
	})
}
