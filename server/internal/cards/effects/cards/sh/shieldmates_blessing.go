package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shieldmate's Blessing — Instant {W} (Zendikar, #5583):
//
//	"Prevent the next 3 damage that would be dealt to any target this
//	 turn."
//
// Mending Hands' charged shield at 3 (CR 615.7): it lasts until end of
// turn and is used up as it prevents. A target gone by resolution gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4ed55e9c-84ea-42a3-a32f-469da56b3eb5",
		Name:         "Shieldmate's Blessing",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetAny(),
		Purpose:      game.Purpose{Answers: game.AnswerPrevent},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.PreventNextDamageToTheTarget("Shieldmate's Blessing", 3)(ctx.Game, item)
		},
	})
}
