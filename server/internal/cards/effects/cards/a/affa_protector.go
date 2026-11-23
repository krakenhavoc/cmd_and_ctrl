package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Affa Protector — Creature — Human Soldier Ally {2}{W}, 1/4 (Oath of the Gatewatch, #5545):
//
//	"Vigilance"
//
// Its whole text is a printed keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5b38a1e4-a6af-4094-9392-70a87759e20d",
		Name:            "Affa Protector",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
