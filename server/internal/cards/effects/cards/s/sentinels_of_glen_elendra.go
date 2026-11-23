package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sentinels of Glen Elendra — Creature — Faerie Soldier {3}{U}, 2/3
// (Jumpstart 2022, #5948):
//
//	"Flash
//	 Flying"
//
// Both keywords ride PrintedKeywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a93779c5-603e-4a91-aef8-a0cd8fd1b3e9",
		Name:            "Sentinels of Glen Elendra",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
	})
}
