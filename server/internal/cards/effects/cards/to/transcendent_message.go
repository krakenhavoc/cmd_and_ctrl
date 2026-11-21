package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Transcendent Message — Instant {X}{U}{U}{U}{U} (March of the Machine):
// "Convoke. Draw X cards."
//
// X is the one the caster announced (CR 107.3); X = 0 draws nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1f7f7765-aeb9-4e47-a6b9-ce823bb29e33",
		Name:         "Transcendent Message",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		TapCost:      effects.Convoke(),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.DrawXCards)),
	})
}
