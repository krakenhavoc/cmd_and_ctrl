package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rally — Instant {W}{W} (Ice Age, S82 OLD, #5636):
//
//	"Blocking creatures get +1/+1 until end of turn."
//
// Morale's pump on blockers: every blocking creature, whoever controls it, as it
// resolves. Outside combat nothing is blocking, so it does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "274f8860-c8c3-4aeb-ba87-94a5139397d8",
		Name:         "Rally",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BoostUntilEOT{Match: effects.BlockingCreature(), Power: 1, Toughness: 1, Label: "Rally — blocking creatures get +1/+1"}.Apply(ctx)
		},
	})
}
