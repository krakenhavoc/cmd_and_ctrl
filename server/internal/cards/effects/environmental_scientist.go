package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Environmental Scientist — Creature — Human Druid {1}{G}, 2/2 (Secrets of
// Strixhaven, #4456):
//
//	"When this creature enters, you may search your library for a basic land
//	 card, reveal it, put it into your hand, then shuffle."
//
// The search is Goblin Matron's: "you may" is the search's own Optional (a
// player may decline, or fail to find, CR 701.23b), the basic land is
// revealed to the table, and the shuffle follows the pick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f30aa897-7027-47e7-ae58-73cbc448c1f0",
		Name:         "Environmental Scientist",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Environmental Scientist — you may search for a basic land card", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player:    item.Controller,
					Predicate: IsBasicLand,
					Dest:      game.ZoneHand,
					Limit:     1,
					Reveal:    true,
					Shuffle:   true,
					Optional:  true,
					Reason:    "Environmental Scientist — a basic land card, revealed, to hand",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
