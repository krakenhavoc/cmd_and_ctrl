package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Shadow Summoning — Sorcery {W}{B} (Foundations Commander):
//
//	"Create two tapped 1/1 white Spirit creature tokens with flying."
//
// The shared 1/1 white Spirit with flying row, created tapped (the
// creation's own property, CreateTokenAdvanced's EntersTapped).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9225046-03aa-4729-89fa-32c049407229",
		Name:         "Shadow Summoning",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateTokenAdvanced{Spec: Token(TokenCard("1/1 white Spirit with flying")).EntersTapped(), N: 2}.Apply(ctx)
		},
	})
}
