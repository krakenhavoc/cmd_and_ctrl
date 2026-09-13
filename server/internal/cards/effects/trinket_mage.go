package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trinket Mage — Creature — Human Wizard {2}{U}, 2/2 (EDHREC rank
// 2038):
//
//	"When this creature enters, you may search your library for an
//	 artifact card with mana value 1 or less, reveal that card, put it
//	 into your hand, then shuffle."
//
// The Sol Ring tutor. An ETB with the S22 search chooser: "you may"
// is the prompt's decline, the filter is artifact cards at mana value
// 0 or 1 (a printed cost read off the library card — Sol Ring, Skullclamp,
// Sensei's Divining Top, and every 0-cost artifact), the pick is
// revealed and goes to hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ffc95093-24a2-4616-a44a-24788e8df9c8",
		Name:         "Trinket Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Trinket Mage — search for an artifact card with mana value 1 or less",
					func(g *game.Game, item *game.StackItem) error {
						return SearchLibrary{
							Player: item.Controller,
							Predicate: func(c game.Card) bool {
								return c.IsArtifact() && c.ManaValue() <= 1
							},
							Dest:     game.ZoneHand,
							Limit:    1,
							Reveal:   true,
							Shuffle:  true,
							Optional: true,
							Reason:   "Trinket Mage — an artifact card with mana value 1 or less",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
