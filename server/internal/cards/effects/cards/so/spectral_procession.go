package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spectral Procession — Sorcery {2/W}{2/W}{2/W} (Shadowmoor, #5591):
//
//	"Create three 1/1 white Spirit creature tokens with flying."
//
// Triplicate Spirits' token row without the convoke: three of the table's
// 1/1 white Spirits with flying. Each {2/W} is paid with {W} or two mana
// (CR 107.4e), and the mana value is 6 (CR 202.3f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "376271a6-80f0-47a6-990a-5ec5517aae42",
		Name:         "Spectral Procession",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.B28WhiteSpiritFlyingToken(), N: 3}),
	})
}
