package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alpine Watchdog — Creature — Dog {1}{W}, 2/2 (Core Set 2021, #5511):
//
//	"Vigilance"
//
// The keyword rides PrintedKeywords; there is no other text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6a0419c3-4bfa-46c1-932b-fe10f2fa2ad7",
		Name:            "Alpine Watchdog",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
