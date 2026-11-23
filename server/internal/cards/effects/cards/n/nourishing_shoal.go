package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nourishing Shoal — Instant — Arcane {X}{G}{G} (Betrayers of Kamigawa,
// #5602):
//
//	"You may exile a green card with mana value X from your hand rather
//	 than pay this spell's mana cost.
//	 You gain X life."
//
// The pitch is Shining Shoal's (bokB1ShoalPitch): X is announced and the
// exiled green card's mana value must equal it. Hard-cast, X is paid as
// usual.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "2aa0e0e5-cb6d-4518-8eff-d29f935486e0",
		Name:             "Nourishing Shoal",
		XMatters:         true,
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.BokB1ShoalPitch("green", "G")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.GainLife{Amount: ctx.X()}.Apply(ctx)
		},
	})
}
