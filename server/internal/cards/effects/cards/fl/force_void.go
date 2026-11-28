package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Force Void — Instant {2}{U} (Ice Age, S82 OLD, #5636):
//
//	"Counter target spell unless its controller pays {1}.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Runeboggle's soft counter with Jolt's delayed draw.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "469bdca3-d373-4420-9266-921b6605f327",
		Name:         "Force Void",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.CounterTheTargetSpellUnlessPaid("Force Void", "{1}")(ctx.Game, item); err != nil {
				return err
			}
			return effects.DrawACardAtTheNextUpkeep("Force Void — draw a card").Apply(ctx)
		},
	})
}
