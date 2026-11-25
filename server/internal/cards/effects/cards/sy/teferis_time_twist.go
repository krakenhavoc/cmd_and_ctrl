package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Teferi's Time Twist — Instant {1}{U} (War of the Spark):
//
//	"Exile target permanent you control. Return that card to the battlefield
//	 under its owner's control at the beginning of the next end step. If it
//	 enters as a creature, it enters with an additional +1/+1 counter on it."
//
// The flicker everyone shares (Mistmeadow Witch's), with its own delayed return
// that adds the counter as the card enters (CR 614.12): the counter is judged on
// what it would be on the battlefield, so a noncreature stays as it was, a token
// never comes back (CR 111.7), and the card returns under its owner's control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f5479d4f-bc4b-4452-a944-41be623f8367",
		Name:         "Teferi's Time Twist",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerSurvive},
		Targets:      effects.TargetPermanent("target permanent you control", effects.YouControl()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileTargetsThenScheduleReturnBody(ctx, "Teferi's Time Twist — return the exiled card", effects.OldB167ReturnExiledCreatureWithCounterBody)
		},
	})
}
