package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Topiary Panther — Creature — Plant Cat {4}{G}{G}, 6/5 (Murders at
// Karlov Manor, #4469):
//
//	"Trample
//	 Basic landcycling {1}{G} ({1}{G}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// Basic landcycling is typecycling for a basic land card (CR 702.29e),
// activated from the hand, as Apex Witchstalker's is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "efddf972-94c4-49d1-a844-b4aba70c93aa",
		Name:            "Topiary Panther",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated:       []effects.ActivatedAbility{effects.BasicLandcycling("{1}{G}")},
	})
}
