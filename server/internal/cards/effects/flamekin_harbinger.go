package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flamekin Harbinger — Creature — Elemental Shaman {R}, 1/1 (Lorwyn, #5593):
//
//	"When this creature enters, you may search your library for an Elemental
//	 card, reveal it, then shuffle and put that card on top."
//
// The Harbinger enters tutor (lrwB1Harbinger): a "you may" asked as the
// trigger resolves (CR 603.5), a search for an Elemental card that may find
// nothing (CR 701.23b), shown to the table, and placed on top after the
// shuffle. A Kindred card of the type counts, and so does a changeling
// card (CR 702.73a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6585e30-4ca0-4701-b274-b24f3508dd97",
		Name:         "Flamekin Harbinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			lrwB1Harbinger("Flamekin Harbinger", "an Elemental card", lrwB1CardOfType("Elemental")),
		},
	})
}
