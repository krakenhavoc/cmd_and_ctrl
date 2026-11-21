package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rampaging Hippo — Creature — Hippo {4}{G}{G}, 5/6
// (Hour of Devastation, #5534):
//
//	"Trample
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The keywords are printed data; cycling is the shared Cycling row, a
// hand ability that discards the card as its cost and draws (CR 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e48aa13b-7488-4644-9080-6c342a8ac463",
		Name:            "Rampaging Hippo",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
