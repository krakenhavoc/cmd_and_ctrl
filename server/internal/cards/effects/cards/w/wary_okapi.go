package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wary Okapi — Creature — Antelope {2}{G}, 3/2 (Guilds of Ravnica, #5525):
//
//	"Vigilance"
//
// Vigilance is an engine keyword carried as printed data (CR 702.20b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "46f01d1c-7ab6-4a58-97ce-57a4eb99843c",
		Name:            "Wary Okapi",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
	})
}
