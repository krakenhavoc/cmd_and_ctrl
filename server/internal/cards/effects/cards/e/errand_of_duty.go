package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Errand of Duty — Instant {1}{W} (Alliances, S82 OLD batch 817):
//
//	"Create a 1/1 white Knight creature token with banding."
func init() {
	effects.Register(effects.Spec{
		OracleID:     "977f0d8d-4359-44ae-8f22-2103ea909c40",
		Name:         "Errand of Duty",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve: effects.ResolveSteps(
			effects.CreateToken{Template: effects.TokenCard("1/1 white Knight with banding"), N: 1},
		),
	})
}
