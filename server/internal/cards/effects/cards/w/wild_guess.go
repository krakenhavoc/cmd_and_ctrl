package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wild Guess — Sorcery {R}{R} (Magic 2013, #5568):
//
//	"As an additional cost to cast this spell, discard a card.
//	 Draw two cards."
//
// Tormenting Voice's shape: the discard is paid at announce (CR
// 601.2h), so it can't be cast from an otherwise empty hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "a05ab59e-680a-4050-aecb-56ee69dcd13f",
		Name:           "Wild Guess",
		Purpose:        game.Purpose{Draws: 2},
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardCost(1),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
	})
}
