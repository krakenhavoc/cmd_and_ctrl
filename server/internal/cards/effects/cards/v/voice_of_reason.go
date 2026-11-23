package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Voice of Reason — Creature — Angel {3}{W}, 2/2 (Urza's Destiny, #5621):
//
//	"Flying, protection from blue"
//
// Voice of Duty with the other colour: no blue creature can block it,
// blue spells and abilities can't target it, and damage from blue
// sources is prevented (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "372e2d6e-427c-4152-b936-27dcf536ca84",
		Name:            "Voice of Reason",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", game.ProtectionFromColor("U")},
	})
}
