package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Needlethorn Drake — Creature — Drake {G}{U}, 1/1 (Strixhaven: School of
// Mages, #5502):
//
//	"Flying, deathtouch"
//
// Two printed keywords the combat engine reads.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6edacc06-9523-4ccf-a4c2-2ac4af386b8a",
		Name:            "Needlethorn Drake",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
	})
}
