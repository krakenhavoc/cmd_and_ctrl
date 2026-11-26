package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Devils' Playground — Sorcery {4}{R}{R} (Shadows over Innistrad, S82 OLD):
//
//	"Create four 1/1 red Devil creature tokens. They have "When this token
//	 dies, it deals 1 damage to any target.""
//
// The Devil is the predefined token (DevilToken), whose printed "When this
// creature dies, it deals 1 damage to any target." is exactly what the card
// gives the tokens (Mogg Alarm's shape).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ab9e2045-860a-40ab-bec1-0dffced003ef",
		Name:         "Devils' Playground",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 4},
		OnResolve:    effects.ResolveSteps(effects.CreateToken{Template: effects.DevilToken(), N: 4}),
	})
}
