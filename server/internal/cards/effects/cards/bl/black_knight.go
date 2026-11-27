package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Black Knight — Creature — Human Knight {B}{B}, 2/2 (Magic 2011, #5579):
//
//	"First strike
//	 Protection from white"
//
// Two printed keywords. Protection from white (CR 702.16) stops white
// spells and abilities from targeting it, white sources from damaging it,
// white Auras from enchanting it and white creatures from blocking it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "9456c5b6-946d-403a-8ed0-dff9f921d98c",
		Name:            "Black Knight",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike", "protection from white"},
	})
}
