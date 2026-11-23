package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Holy Day — Instant {W}:
//
//	"Prevent all combat damage that would be dealt this turn."
//
// Fog in white, word for word. Registered separately because the
// catalog keys on oracle ID and a white deck runs this one; the
// effect is the shared primitive.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98423a34-f044-4811-b288-56981d604b6e",
		Name:         "Holy Day",
		Purpose:      game.Purpose{Answers: game.AnswerPrevent}, // #4734 slice 6
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.PreventAllCombatDamageThisTurn{
				Label: "Holy Day: prevent combat damage",
			}.Apply(ctx)
		},
	})
}
