package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Meeting of Minds — Instant {3}{U} (March of the Machine): "Convoke. Draw
// two cards."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3b91dc3e-46ef-4a68-829f-ef7d79598cf7",
		Name:         "Meeting of Minds",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      game.Purpose{Draws: 2},
		OnResolve:    effects.ResolveSteps(effects.Instruction(func(ctx *effects.Context) error { return effects.DrawCards{N: 2}.Apply(ctx) })),
	})
}
