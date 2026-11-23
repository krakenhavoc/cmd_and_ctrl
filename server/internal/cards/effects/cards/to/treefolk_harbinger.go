package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Treefolk Harbinger — Creature — Treefolk Druid {G}, 0/3 (Lorwyn, #5593):
//
//	"When this creature enters, you may search your library for a Treefolk
//	 or Forest card, reveal it, then shuffle and put that card on top."
//
// The Harbinger enters tutor (lrwB1Harbinger): a "you may" asked as the
// trigger resolves (CR 603.5), a search for a Treefolk or Forest card that may find
// nothing (CR 701.23b), shown to the table, and placed on top after the
// shuffle. A Kindred card of the type counts, and so does a changeling
// card (CR 702.73a).
//
// A Forest card is any land card with the Forest subtype, a nonbasic one
// included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "98c387bd-c9ba-464a-8380-839a5604a37f",
		Name:         "Treefolk Harbinger",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.LrwB1Harbinger("Treefolk Harbinger", "a Treefolk or Forest card", func(c game.Card) bool {
				return effects.LrwB1CardOfType("Treefolk")(c) || effects.IsLandWithSubtype("forest")(c)
			}),
		},
	})
}
