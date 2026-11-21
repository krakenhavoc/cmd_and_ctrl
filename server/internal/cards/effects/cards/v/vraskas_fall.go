package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vraska's Fall — Instant {2}{B}: "Each opponent sacrifices a creature or
// planeswalker of their choice and gets a poison counter."
//
// The edict asks each opponent; the poison counters are its keyed
// continuation, run once every opponent has sacrificed. Every opponent gets
// the counter, whether or not they had anything to sacrifice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "33fd9192-3e9b-4d3a-92aa-06de29d449df",
		Name:         "Vraska's Fall",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.EachPlayerSacrifices{
				ExceptController: true,
				Match:            effects.Or(effects.Creature(), effects.Planeswalker()),
				Label:            "a creature or planeswalker",
				Keyed:            effects.S82OneB04Edicted,
			}.Apply(ctx)
		},
	})
}
