package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "ee05fc58-1356-4b2f-b9ce-bda8cc716aee",
		Name:            "Augury Raven",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		SpecialActions:  []game.SpecialAction{effects.Foretell("{1}{U}")},
	})
}
