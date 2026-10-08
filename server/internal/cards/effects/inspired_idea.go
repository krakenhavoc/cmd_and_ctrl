package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inspired Idea — Sorcery {2}{U}:
//
//	"Cleave {3}{U}{U} (You may cast this spell for its cleave cost. If
//	 you do, remove the words in square brackets.)
//	 Draw three cards. [Your maximum hand size is reduced by three for
//	 the rest of the game.]"
//
// The reduction is a granted maximum-hand-size change with a duration
// (game.GrantHandSizeForEffect, ADR 0113's amendment of 2026-10-08,
// #2108): a Modify of -3 that lives for the rest of the game and is
// folded in CR 613.11 timestamp order with every other change. A second
// Inspired Idea reduces by three again. Cleaving removes the bracketed
// sentence, so the cleaved spell is a plain draw three. The brackets do
// not touch the targeting clause (there is none), hence Cleave's nil.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "5f1386d5-b802-46b2-9803-275621ebda80",
		Name:             "Inspired Idea",
		Completeness:     CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{Cleave("{3}{U}{U}", nil)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if err := (DrawCards{Player: item.Controller, N: 3}).Apply(ctx); err != nil {
				return err
			}
			if ctx.PaidAltCost("cleave") {
				return nil
			}
			return ctx.Game.GrantHandSizeForEffect(item.Controller, game.HandSizeModify, -3,
				"Inspired Idea — maximum hand size reduced by three", item.SourceCardID, game.IndefiniteDuration())
		},
	})
}
