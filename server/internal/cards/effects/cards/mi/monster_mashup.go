package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monster Mashup — {3}{R} Creature — Werewolf Fish Zombie Vampire 4/3
// (Teenage Mutant Ninja Turtles Commander, #5011):
//
//	"Reach
//	 Menace"
//
// Two printed keywords, both engine-read (CR 702.17, CR 702.111).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "485e1d84-2867-4c66-a399-cdc2a22f6fc9",
		Name:            "Monster Mashup",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "menace"},
	})
}
