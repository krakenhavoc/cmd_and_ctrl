package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Heal — Instant {W} (Ice Age, S82 OLD, #5636):
//
//	"Prevent the next 1 damage that would be dealt to any target this turn.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Benevolent Ancestor's shield as a spell, then Jolt's delayed draw
// (CR 615.7, 603.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "013c42a4-d555-46c9-9d57-b50d6b59ae9e",
		Name:         "Heal",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPrevent, Draws: 1},
		Targets:      effects.TargetAny(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.PreventNextOneDamageToTheTarget("Heal")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Heal — draw a card").Apply(ctx)
		},
	})
}
