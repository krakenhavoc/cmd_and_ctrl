package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sworn Companions — Sorcery {2}{W} (Guilds of Ravnica, #5525):
//
//	"Create two 1/1 white Soldier creature tokens with lifelink."
//
// The token table's 1/1 white lifelink Soldier, twice.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d1a9fd14-fecb-4d2a-983c-70143238275b",
		Name:         "Sworn Companions",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2, CreatureTokens: 2},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 white Soldier with lifelink"), N: 2}),
	})
}
