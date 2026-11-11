package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Imperial Hellkite — Creature — Dragon {5}{R}{R}, 6/6:
//
//	"Flying
//	 Morph {6}{R}{R}
//	 When this creature is turned face up, you may search your library
//	 for a Dragon card, reveal it, put it into your hand, then shuffle."
//
// Brightglass Gearhulk's tutor with a Dragon filter and a count of one.
// Declining skips the search and the shuffle.
//
// No simplification.
func init() {
	const label = "Imperial Hellkite — search for a Dragon card"
	Register(Spec{
		OracleID:         "99642419-ab24-4770-9f1c-84d2806441b5",
		Name:             "Imperial Hellkite",
		Completeness:     CompletenessFull,
		PrintedKeywords:  []string{"flying"},
		AlternativeCosts: []game.AlternativeCost{Morph("{6}{R}{R}")},
		Triggered: []game.TriggeredAbility{
			WhenThisIsTurnedFaceUp(label, func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: func(c game.Card) bool { return c.HasSubtypeWord("Dragon") },
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    label,
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
