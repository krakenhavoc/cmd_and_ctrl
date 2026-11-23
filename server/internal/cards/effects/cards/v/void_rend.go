package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Void Rend — Instant {W}{U}{B}:
//
//	"This spell can't be countered.
//	 Destroy target nonland permanent."
//
// Abrupt Decay without the mana-value cap, at triple the color
// commitment. Same two pieces: Spec.CantBeCountered (S23) and a
// Nonland() target clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "713f16db-95ec-479e-a48c-7a69f7668d7f",
		Name:            "Void Rend",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Answers: game.AnswerRemove}, // #4734: destroys a permanent
		CantBeCountered: true,
		Targets:         effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyChosenPermanent(ctx.Game, item)
		},
	})
}
