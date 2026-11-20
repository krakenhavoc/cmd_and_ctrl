package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Allied Reinforcements — Sorcery {3}{W} (Oath of the Gatewatch, #5545):
//
//	"Create two 2/2 white Knight Ally creature tokens."
//
// The Knight Ally is the token table's 2/2 white Knight Ally.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2baf2ee6-e290-4a2b-89a0-0e725edde06d",
		Name:         "Allied Reinforcements",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve:    ResolveSteps(CreateToken{Template: TokenCard("2/2 white Knight Ally"), N: 2}),
	})
}
