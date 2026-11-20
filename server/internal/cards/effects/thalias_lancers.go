package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thalia's Lancers — Creature — Human Knight {3}{W}{W}, 4/4 (Eldritch
// Moon, #5541):
//
//	"First strike
//	 When this creature enters, you may search your library for a
//	 legendary card, reveal it, put it into your hand, then shuffle."
//
// Heliod's Pilgrim's tutor with Captain Sisay's legendary predicate. "You
// may search" is the search's own Optional (#3053).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "58b349ec-9987-4832-a8f3-8e92027ea75b",
		Name:            "Thalia's Lancers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Purpose:         game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Thalia's Lancers — you may search for a legendary card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.IsLegendary() },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Thalia's Lancers — a legendary card",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
