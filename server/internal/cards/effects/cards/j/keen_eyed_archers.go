package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Keen-Eyed Archers — Creature — Elf Archer {2}{W}, 2/2 (Portal, #5630):
//
//	"Reach"
//
// Reach is an engine keyword carried as printed data: it can block
// creatures with flying (CR 702.17b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0ace32d6-7261-447c-9ee2-e03febaab91b",
		Name:            "Keen-Eyed Archers",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
