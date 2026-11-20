package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Battalion Foot Soldier — Creature — Human Soldier {2}{W}, 2/2 (Core Set 2020, #5518):
//
//	"When this creature enters, you may search your library for any number
//	 of cards named Battalion Foot Soldier, reveal them, put them into your
//	 hand, then shuffle."
//
// Andromedan Automatons' search: SearchLibrary's Unbounded pick (zero up
// to every copy), and "you may" lets the controller skip the search and
// the shuffle.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f6f39a85-28b9-49f3-ab90-2ff99169de42",
		Name:         "Battalion Foot Soldier",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Battalion Foot Soldier — search for cards named Battalion Foot Soldier",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.Name == "Battalion Foot Soldier" },
						Dest:      game.ZoneHand,
						Unbounded: true,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Battalion Foot Soldier — you may search for any number of cards named Battalion Foot Soldier",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
