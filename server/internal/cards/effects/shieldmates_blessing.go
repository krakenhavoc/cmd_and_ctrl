package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "4ed55e9c-84ea-42a3-a32f-469da56b3eb5",
		Name:         "Shieldmate's Blessing",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		Purpose:      game.Purpose{Answers: game.AnswerPrevent},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return preventNextDamageToTheTarget("Shieldmate's Blessing", 3)(ctx.Game, item)
		},
	})
}
