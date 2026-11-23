package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Contact Other Plane — Instant {3}{U} (Adventures in the Forgotten
// Realms, #5499):
//
//	"Roll a d20.
//	 1—9 | Draw two cards.
//	 10—19 | Scry 2, then draw two cards.
//	 20 | Scry 3, then draw three cards."
//
// The draw after a scry waits for the scry to finish (Preordain's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a352cc67-eae8-4fa5-abba-507821c4cf47",
		Name:         "Contact Other Plane",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 2},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.RollDice{Sides: 20, N: 1, Then: contactOtherPlaneTable}.Apply(ctx)
		},
	})
}

// contactOtherPlaneTable is the results table, after the d20.
var contactOtherPlaneTable = effects.DiceTail("dice/contact-other-plane-table", func(ctx *effects.Context, rolls []int) error {
	switch {
	case rolls[0] >= 20:
		return effects.Scry{N: 3, ThenTail: ctx.TailN(effects.ThenDrawCards, 3)}.Apply(ctx)
	case rolls[0] >= 10:
		return effects.Scry{N: 2, ThenTail: ctx.TailN(effects.ThenDrawCards, 2)}.Apply(ctx)
	}
	return effects.DrawCards{N: 2}.Apply(ctx)
})
