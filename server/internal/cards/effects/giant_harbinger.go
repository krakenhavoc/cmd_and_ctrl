package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Giant Harbinger — Creature — Giant Shaman {4}{R}, 3/4 (Lorwyn, #5593):
//
//	"When this creature enters, you may search your library for a Giant
//	 card, reveal it, then shuffle and put that card on top."
//
// The Harbinger enters tutor (lrwB1Harbinger): a "you may" asked as the
// trigger resolves (CR 603.5), a search for a Giant card that may find
// nothing (CR 701.23b), shown to the table, and placed on top after the
// shuffle. A Kindred card of the type counts, and so does a changeling
// card (CR 702.73a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2445e58b-87ed-4ab2-8209-a5e1f566fba7",
		Name:         "Giant Harbinger",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			lrwB1Harbinger("Giant Harbinger", "a Giant card", lrwB1CardOfType("Giant")),
		},
	})
}
