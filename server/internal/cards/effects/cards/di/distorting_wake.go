package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Distorting Wake — Sorcery {X}{U}{U}{U} (Commander 2014):
//
//	"Return X target nonland permanents to their owners' hands."
//
// Exactly X targets (CountFromX), each re-checked as the spell resolves
// (CR 608.2b); the ones still legal go home.
func init() {
	spec := effects.TargetPermanent("X target nonland permanents", effects.Nonland())
	spec.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "e5387fc7-c3b3-4933-bffa-580ad793c7ab",
		Name:         "Distorting Wake",
		Completeness: effects.CompletenessFull,
		Targets:      spec,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.BounceEachLegalTarget(ctx.Game, item)
		},
	})
}
