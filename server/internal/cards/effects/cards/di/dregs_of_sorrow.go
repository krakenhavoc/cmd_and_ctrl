package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dregs of Sorrow — Sorcery {X}{4}{B} (Commander 2014):
//
//	"Destroy X target nonblack creatures. Draw X cards."
//
// Exactly X targets (CountFromX), each judged again as it resolves
// (CR 608.2b); the cards are drawn for X whatever the number destroyed,
// but only when at least one target is still legal (a spell with all its
// targets illegal does not resolve).
func init() {
	spec := effects.TargetCreature("X target nonblack creatures", effects.NonBlack())
	spec.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "8126e60b-cf3d-4bf0-bfb5-f1e914d3e20e",
		Name:         "Dregs of Sorrow",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      spec,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.DestroyEachLegalTarget(item, ctx); err != nil {
				return err
			}
			return effects.DrawCards{Player: item.Controller, N: ctx.X()}.Apply(ctx)
		},
	})
}
