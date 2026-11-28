package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Headstone — Instant {1}{B} (Homelands, S82 OLD b827, #5635):
//
//	"Exile target card from a graveyard.
//	 Draw a card at the beginning of the next turn's upkeep."
//
// Moratorium Stone's graveyard exile with Gravebind's delayed draw: a CR 603.7
// delayed trigger at the next upkeep at the table, whoever's it is, controlled
// by Headstone's controller. A target that is no longer legal takes the whole
// spell with it (CR 608.2b), the delayed draw included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cbf767d4-9be4-417a-bbf7-0c72ede727b4",
		Name:         "Headstone",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Targets:      effects.TargetCardInGraveyard("target card from a graveyard"),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(ctx.LegalTargets()) == 0 {
				return nil
			}
			if err := effects.ExileTargetCards(ctx.Game, item); err != nil {
				return err
			}
			return effects.ScheduleDelayedTrigger{
				At:    game.StepUpkeep,
				Label: "Headstone — draw a card",
				Body:  effects.DrawOneBody,
			}.Apply(ctx)
		},
	})
}
