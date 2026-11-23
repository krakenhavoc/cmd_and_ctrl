package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flight of Equenauts — Creature — Human Knight {7}{W}, 4/5:
//
//	"Convoke
//	 Flying"
//
// Convoke is Spec.TapCost, as Ancient Imperiosaur has it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5212e707-b487-46b6-aa51-cea3be8cbe00",
		Name:            "Flight of Equenauts",
		Completeness:    effects.CompletenessFull,
		TapCost:         effects.Convoke(),
		PrintedKeywords: []string{"flying"},
	})
}
