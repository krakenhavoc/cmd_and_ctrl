package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shu General — Creature — Human Soldier {3}{W}, 2/2
// (Portal Three Kingdoms, #5622):
//
//	"Vigilance; horsemanship (This creature can't be blocked except by creatures with horsemanship.)"
//
// Vigilance is an engine keyword carried as printed data: attacking
// doesn't tap it (CR 702.20b).
//
// Horsemanship is an engine keyword carried as printed data: the block
// check refuses any blocker without horsemanship. It doesn't stop this
// creature blocking.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0fef6729-b044-4daf-834a-ac8c9711ad8a",
		Name:            "Shu General",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance", "horsemanship"},
	})
}
