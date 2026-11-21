package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Triplicate Spirits — Sorcery {4}{W}{W} (Foundations Jumpstart):
//
//	"Convoke (Your creatures can help cast this spell. Each creature you
//	 tap while casting this spell pays for {1} or one mana of that
//	 creature's color.)
//	 Create three 1/1 white Spirit creature tokens with flying."
//
// Convoke is the shared TapCost; the body is three of the existing
// 1/1 white flying Spirit row, made together.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "48090c64-f41a-447b-ad7a-54e3194f759b",
		Name:         "Triplicate Spirits",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Purpose:      game.Purpose{Tokens: 3, CreatureTokens: 3},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.TokenCard("1/1 white Spirit with flying"), N: 3}),
	})
}
