package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Into the Core — Instant {2}{R}{R} (Mirrodin Besieged,
// #5576):
//
//	"Exile two target artifacts."
//
// Dust to Dust's exactly-two clause at instant speed: each target still
// legal as it resolves is exiled, so one that left in response is skipped
// and the other is still exiled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ad08ad3a-bc58-43da-b7f3-4a478132da8a",
		Name:         "Into the Core",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // #4734: exiles permanents
		Targets:      effects.TargetPermanent("two target artifacts", effects.Artifact()).WithCount(2, 2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ExileEveryLegalCardTarget(ctx.Game, item)
		},
	})
}
