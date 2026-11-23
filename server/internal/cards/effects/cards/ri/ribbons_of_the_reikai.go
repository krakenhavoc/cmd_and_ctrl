package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ribbons of the Reikai — Sorcery — Arcane {4}{U} (Betrayers of Kamigawa,
// #5602):
//
//	"Draw a card for each Spirit you control."
//
// The count is the Spirits the caster controls as the spell resolves (CR
// 608.2h), by effective subtype, so a changeling counts and "Spirit"
// reads through a text change.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22b76ab7-9658-41a3-b8bb-fc02de0631a4",
		Name:         "Ribbons of the Reikai",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerValue},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			spirit := ctx.TextSubtype("Spirit")
			n := effects.CountControlled(ctx.Game, ctx.Controller(), func(c game.Card) bool { return c.HasSubtypeWord(spirit) })
			return effects.DrawCards{N: n}.Apply(ctx)
		},
	})
}
