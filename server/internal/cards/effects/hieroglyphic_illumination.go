package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "60e358ec-eb11-4a16-a0ae-b7c064a8ea6c",
		Name:         "Hieroglyphic Illumination",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
		},
		Activated: []ActivatedAbility{Cycling("{U}")},
	})
}
