package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Faerie Harbinger — Creature — Faerie Wizard {3}{U}, 2/2 (Lorwyn, #5593):
//
//	"Flash
//	 Flying
//	 When this creature enters, you may search your library for a Faerie
//	 card, reveal it, then shuffle and put that card on top."
//
// The Harbinger enters tutor (lrwB1Harbinger): a "you may" asked as the
// trigger resolves (CR 603.5), a search for a Faerie card that may find
// nothing (CR 701.23b), shown to the table, and placed on top after the
// shuffle. A Kindred card of the type counts, and so does a changeling
// card (CR 702.73a).
//
// Flash (CR 702.8a) and flying (CR 702.9b) are printed keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "a171a01c-43f7-488a-870d-b1a9bcc501d1",
		Name:            "Faerie Harbinger",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.LrwB1Harbinger("Faerie Harbinger", "a Faerie card", effects.LrwB1CardOfType("Faerie")),
		},
	})
}
