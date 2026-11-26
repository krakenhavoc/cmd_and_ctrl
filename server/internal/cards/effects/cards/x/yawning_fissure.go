package cardsx

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yawning Fissure — Sorcery {4}{R} (Magic 2010, #5585):
//
//	"Each opponent sacrifices a land of their choice."
//
// Sheoldred's Edict's fan-out over the opponents in APNAP order, each
// choosing a land; nothing is targeted.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "634ec9d3-24c9-4090-a087-2624b6d5bf5b",
		Name:         "Yawning Fissure",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EachPlayerSacrifices{ExceptController: true, Match: effects.Land(), Label: "a land"}.Apply(ctx)
		},
	})
}
