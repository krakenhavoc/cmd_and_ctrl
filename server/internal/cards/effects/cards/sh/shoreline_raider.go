package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Shoreline Raider — Creature — Merfolk {2}{U}, 2/2 (Invasion, S82 #5616):
//
//	"Protection from Kavu"
//
// Protection from a creature type is a printed keyword: it applies to every Kavu, whoever controls it (CR 702.16).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "79dcc432-e8f5-4f2b-bff8-5ae964714666",
		Name:            "Shoreline Raider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"protection from Kavu"},
	})
}
