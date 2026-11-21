package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Archweaver — Creature — Spider {5}{G}{G}, 5/5 (Return to Ravnica, #5566):
//
//	"Reach, trample"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2ceb76e2-1692-48fc-ab20-9587d1374172",
		Name:            "Archweaver",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach", "trample"},
	})
}
