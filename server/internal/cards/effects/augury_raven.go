package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Augury Raven — Creature — Bird {3}{U}, 3/3 (Kaldheim, #5505):
//
//	"Flying
//	 Foretell {1}{U}"
//
// The keywords ride PrintedKeywords; foretell is the shared special
// action (foretell.go, CR 702.143), declared at the printed cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ee05fc58-1356-4b2f-b9ce-bda8cc716aee",
		Name:            "Augury Raven",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{Foretell("{1}{U}")},
	})
}
