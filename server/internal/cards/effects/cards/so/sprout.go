package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sprout — Instant {G} (Time Spiral, #5596):
//
//	"Create a 1/1 green Saproling creature token."
//
// One 1/1 green Saproling token for its caster, at instant speed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "424c6f5e-b386-47e9-b3fe-25b263097d40",
		Name:         "Sprout",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateToken{Template: effects.TokenCard("1/1 green Saproling"), N: 1}.Apply(ctx)
		},
	})
}
