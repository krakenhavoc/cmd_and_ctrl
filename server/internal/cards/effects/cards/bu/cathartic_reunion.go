package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cathartic Reunion — Sorcery {1}{R}:
//
//	"As an additional cost to cast this spell, discard two cards.
//	 Draw three cards."
//
// The two-for-three rate the discard-and-draw cycle scales to at
// {1}{R}. Same cost shape as Thrill of Possibility, just N=2.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "0f3c3e5f-6af3-4af2-8703-4ccc8ed8f675",
		Name:           "Cathartic Reunion",
		Completeness:   effects.CompletenessFull,
		AdditionalCost: effects.DiscardCost(2),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 3}.Apply(ctx)
		},
	})
}
