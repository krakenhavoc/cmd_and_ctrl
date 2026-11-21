package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Call of the Conclave — Sorcery {G}{W} (Return to Ravnica, #5566):
//
//	"Create a 3/3 green Centaur creature token."
//
// The token is the table's 3/3 green Centaur row (Centaur Glade).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e1635acd-ed1e-4038-a11a-6518df285253",
		Name:         "Call of the Conclave",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("3/3 green Centaur"), N: 1}),
	})
}
