package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Imperial Oath — Sorcery {5}{W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Create three 2/2 white Samurai creature tokens with vigilance. Scry 3."
//
// The tokens are made together, then the scry (CR 701.22) is asked.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4a4a061d-bc4e-49da-8d20-90fcb571b025",
		Name:         "Imperial Oath",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Scry: 3, Answers: game.AnswerValue},
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.NeoB08SamuraiToken(), N: 3},
			effects.Scry{N: 3},
		),
	})
}
