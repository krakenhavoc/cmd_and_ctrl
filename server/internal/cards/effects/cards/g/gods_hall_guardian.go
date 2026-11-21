package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gods' Hall Guardian — Creature — Cat {5}{W}, 3/6 (Kaldheim, #5505):
//
//	"Vigilance
//	 Foretell {3}{W}"
//
// The keywords ride PrintedKeywords; foretell is the shared special
// action (foretell.go, CR 702.143), declared at the printed cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7d22e071-4af4-4182-b387-ecca873d00cc",
		Name:            "Gods' Hall Guardian",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		SpecialActions:  []game.SpecialAction{effects.Foretell("{3}{W}")},
	})
}
