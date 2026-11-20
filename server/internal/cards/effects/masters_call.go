package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Master's Call — Instant {2}{W} (Mirrodin Besieged,
// #5576):
//
//	"Create two 1/1 colorless Myr artifact creature tokens."
//
// Memnarch's two Myr from the token table as an instant: the tokens are
// created under the caster's control.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6c9c1f0-050f-40e2-ab6c-6566d0af418c",
		Name:         "Master's Call",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2, Answers: game.AnswerMakesBlocker},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("1/1 colorless Myr artifact"), N: 2}),
	})
}
