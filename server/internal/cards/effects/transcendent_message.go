package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Transcendent Message — Instant {X}{U}{U}{U}{U} (March of the Machine):
// "Convoke. Draw X cards."
//
// X is the one the caster announced (CR 107.3); X = 0 draws nothing.
func init() {
	Register(Spec{
		OracleID:     "1f7f7765-aeb9-4e47-a6b9-ce823bb29e33",
		Name:         "Transcendent Message",
		Completeness: CompletenessFull,
		XMatters:     true,
		TapCost:      Convoke(),
		Purpose:      game.Purpose{Draws: 1},
		OnResolve:    ResolveSteps(Instruction(drawXCards)),
	})
}
