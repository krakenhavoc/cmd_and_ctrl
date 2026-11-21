package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sporecap Spider — Creature — Spider {2}{G}, 1/5 (Rise of the Eldrazi,
// #5581):
//
//	"Reach"
//
// A keyword-only body (see giant_spider.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e9b5d4f3-b98a-463a-b628-289c2df81204",
		Name:            "Sporecap Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
