package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ranger of Eos — Creature — Human Soldier Ranger {3}{W}, 3/2 (Shards of
// Alara, #5588):
//
//	"When this creature enters, you may search your library for up to two
//	 creature cards with mana value 1 or less, reveal them, put them into
//	 your hand, then shuffle."
//
// Ranger-Captain of Eos's search with a limit of two: the search has a
// stated quality (CR 701.23a), so it may find fewer, and the "you may"
// lets the searcher decline the search and the shuffle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9c8b929-f3ee-457f-9cbc-9de18eb69dc5",
		Name:         "Ranger of Eos",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tutors: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Ranger of Eos — search for up to two creature cards with mana value 1 or less", func(g *game.Game, item *game.StackItem) error {
				return SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.IsCreature() && c.ManaValue() <= 1
					},
					Dest:     game.ZoneHand,
					Limit:    2,
					Reveal:   true,
					Shuffle:  true,
					Optional: true,
					Reason:   "Ranger of Eos — up to two creature cards with mana value 1 or less",
				}.Apply(NewContext(g, item))
			}),
		},
	})
}
