package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperial Oath — Sorcery {5}{W} (Kamigawa: Neon Dynasty, #5493):
//
//	"Create three 2/2 white Samurai creature tokens with vigilance. Scry 3."
//
// The tokens are made together, then the scry (CR 701.22) is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4a4a061d-bc4e-49da-8d20-90fcb571b025",
		Name:         "Imperial Oath",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Scry: 3, Answers: game.AnswerValue},
		OnResolve: ResolveSteps(
			CreateToken{Template: neoB08SamuraiToken(), N: 3},
			Scry{N: 3},
		),
	})
}
