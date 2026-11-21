package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hieroglyphic Illumination — Instant {3}{U}
// (Amonkhet, #5535):
//
//	"Draw two cards.
//	 Cycling {U} ({U}, Discard this card: Draw a card.)"
//
// Draw two, plus cycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "60e358ec-eb11-4a16-a0ae-b7c064a8ea6c",
		Name:         "Hieroglyphic Illumination",
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{U}")},
	})
}
