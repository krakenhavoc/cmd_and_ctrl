package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spectral Reserves — Sorcery {3}{W} (Eldritch Moon, #5541):
//
//	"Create two 1/1 white Spirit creature tokens with flying. You gain 2
//	 life."
//
// Midnight Haunting's two Spirits, then the life, in printed order
// (CR 608.2c).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "628e5c1d-c7f9-4a98-98fb-13d7d5f81f64",
		Name:         "Spectral Reserves",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, Answers: game.AnswerMakesBlocker},
		OnResolve: ResolveSteps(
			CreateToken{Template: TokenCard("1/1 white Spirit with flying"), N: 2},
			GainLife{Amount: 2},
		),
	})
}
