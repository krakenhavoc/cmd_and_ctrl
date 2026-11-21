package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mental Journey — Instant {4}{U}{U} (Modern Horizons 2, #5501):
//
//	"Draw three cards.
//	 Basic landcycling {1}{U} ({1}{U}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// Concentrate's draw at instant speed, and the cycling family's
// BasicLandcycling (CR 702.29e), which works only from hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "94e71162-7c9d-46f4-90e8-413a1b08a198",
		Name:         "Mental Journey",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Draws: 3},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: ctx.Controller(), N: 3}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.BasicLandcycling("{1}{U}")},
	})
}
