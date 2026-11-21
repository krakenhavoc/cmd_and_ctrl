package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Downpour — Instant {1}{U} (Magic 2013, #5568):
//
//	"Tap up to three target creatures."
//
// Zero to three creature targets; each one still legal as it resolves
// is tapped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78eb5fb1-cb91-4300-9056-4225a416aec0",
		Name:         "Downpour",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("up to three target creatures").WithCount(0, 3),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.TapEachLegalTarget(ctx.Game, item)
		},
	})
}
