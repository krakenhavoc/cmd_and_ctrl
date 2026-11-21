package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Bloodscent — Instant, {3}{G}:
//
//	"All creatures able to block target creature this turn do so."
//
// Alluring Scent at instant speed (#1684) — cast after attackers are
// declared, it turns a combat the defender had already planned.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b2aabfaf-5996-4b41-987a-6bd941f6ae88",
		Name:         "Bloodscent",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.LureTargetForTheTurn(ctx, "Bloodscent — all creatures able to block it do so")
		},
	})
}
