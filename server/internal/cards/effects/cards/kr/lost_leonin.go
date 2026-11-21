package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lost Leonin — Creature — Phyrexian Cat Soldier {1}{W}, 2/1 (New Phyrexia, #5575):
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1 counters and to players in the form of poison counters.)"
//
// Infect is the engine's damage tail (CR 702.90).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c21b5c5b-9191-4748-b111-a51631e0bc5a",
		Name:            "Lost Leonin",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
	})
}
