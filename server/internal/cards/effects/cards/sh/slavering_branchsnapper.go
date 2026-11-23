package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Slavering Branchsnapper — Creature — Lizard {4}{G}{G}, 7/6 (Duskmourn:
// House of Horror, #4466):
//
//	"Trample
//	 Forestcycling {2} ({2}, Discard this card: Search your library for
//	 a Forest card, reveal it, put it into your hand, then shuffle.)"
//
// Forestcycling is the shared Typecycling constructor (CR 702.29e), as on
// Daggermaw Megalodon: any card with the Forest subtype, basic or not.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "29326487-6d95-40a8-9bcc-f94ada4e27c3",
		Name:            "Slavering Branchsnapper",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{
			effects.Typecycling("Forestcycling", "{2}", "a Forest card", effects.IsLandWithSubtype("Forest")),
		},
	})
}
