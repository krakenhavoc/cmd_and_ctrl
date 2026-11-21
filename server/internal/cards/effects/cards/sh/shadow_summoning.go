package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shadow Summoning — Sorcery {W}{B} (Foundations Commander):
//
//	"Create two tapped 1/1 white Spirit creature tokens with flying."
//
// The shared 1/1 white Spirit with flying row, created tapped (the
// creation's own property, CreateTokenAdvanced's EntersTapped).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9225046-03aa-4729-89fa-32c049407229",
		Name:         "Shadow Summoning",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateTokenAdvanced{Spec: effects.Token(effects.TokenCard("1/1 white Spirit with flying")).EntersTapped(), N: 2}.Apply(ctx)
		},
	})
}
