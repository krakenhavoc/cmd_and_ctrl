package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boggart Harbinger — Creature — Goblin Shaman {2}{B}, 2/1 (Lorwyn, #5593):
//
//	"When this creature enters, you may search your library for a Goblin
//	 card, reveal it, then shuffle and put that card on top."
//
// The Harbinger enters tutor (lrwB1Harbinger): a "you may" asked as the
// trigger resolves (CR 603.5), a search for a Goblin card that may find
// nothing (CR 701.23b), shown to the table, and placed on top after the
// shuffle. A Kindred card of the type counts, and so does a changeling
// card (CR 702.73a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "56508ae7-6013-4149-8935-96b8d5091baa",
		Name:         "Boggart Harbinger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LrwB1Harbinger("Boggart Harbinger", "a Goblin card", effects.LrwB1CardOfType("Goblin")),
		},
	})
}
