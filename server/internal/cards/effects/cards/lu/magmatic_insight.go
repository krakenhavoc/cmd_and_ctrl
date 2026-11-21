package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magmatic Insight — Sorcery {R}:
//
//	"As an additional cost to cast this spell, discard a land card.
//	 Draw two cards."
//
// The discard is a cost (CR 601.2h), paid with the spell on the stack,
// so a discard payoff sees it before the cards are drawn and a countered
// Insight keeps nothing. Only a land card pays it: with none in hand the
// spell cannot be cast (CR 118.3, game.AdditionalCost.DiscardLand).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "6a9031b6-3f4d-4a36-90e1-ac77b8518cee",
		Name:           "Magmatic Insight",
		Purpose:        game.Purpose{Draws: 2},
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardLandCost(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
