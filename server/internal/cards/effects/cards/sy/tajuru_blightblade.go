package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tajuru Blightblade — Creature — Elf Rogue {G}, 1/1 (Zendikar Rising,
// #5508):
//
//	"Deathtouch"
//
// An engine keyword carried as printed data (CR 702.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "debbbdd8-c8e2-4384-bcac-dbdfb18b54fa",
		Name:            "Tajuru Blightblade",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
