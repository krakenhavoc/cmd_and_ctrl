package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Merrow Harbinger — Creature — Merfolk Wizard {3}{U}, 2/3 (Lorwyn, #5593):
//
//	"Islandwalk (This creature can't be blocked as long as defending player
//	 controls an Island.)
//	 When this creature enters, you may search your library for a Merfolk
//	 card, reveal it, then shuffle and put that card on top."
//
// The Harbinger enters tutor (lrwB1Harbinger): a "you may" asked as the
// trigger resolves (CR 603.5), a search for a Merfolk card that may find
// nothing (CR 701.23b), shown to the table, and placed on top after the
// shuffle. A Kindred card of the type counts, and so does a changeling
// card (CR 702.73a).
//
// Islandwalk is the engine's landwalk keyword (CR 702.14c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "035ee753-232f-4ba2-afbf-8d1eb162e6f8",
		Name:            "Merrow Harbinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"islandwalk"},
		Triggered: []game.TriggeredAbility{
			effects.LrwB1Harbinger("Merrow Harbinger", "a Merfolk card", effects.LrwB1CardOfType("Merfolk")),
		},
	})
}
