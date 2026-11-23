package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Alert Shu Infantry — Creature — Human Soldier {2}{W}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"Vigilance"
//
// Vigilance is an engine keyword carried as printed data: attacking
// doesn't tap it (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0f019387-4684-493f-80c4-de6023096306",
		Name:            "Alert Shu Infantry",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
