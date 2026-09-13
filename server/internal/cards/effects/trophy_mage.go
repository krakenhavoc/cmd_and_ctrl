package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Trophy Mage — Creature — Human Wizard {2}{U}, 2/2 (EDHREC rank
// 2295):
//
//	"When this creature enters, you may search your library for an
//	 artifact card with mana value 3, reveal it, put it into your
//	 hand, then shuffle."
//
// The middle Mage: Trinket Mage's ETB with the filter at exactly
// three (a printed cost read off the library card — Basalt
// Monolith, Coalition Relic, Crucible of Worlds). "You may" is the
// search prompt's decline; the pick is revealed and goes to hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "08afc0d7-192f-4ab6-b6a0-c4265cf5e225",
		Name:         "Trophy Mage",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Trophy Mage — search for an artifact card with mana value 3",
					func(g *game.Game, item *game.StackItem) error {
						return SearchLibrary{
							Player: item.Controller,
							Predicate: func(c game.Card) bool {
								return c.IsArtifact() && c.ManaValue() == 3
							},
							Dest:     game.ZoneHand,
							Limit:    1,
							Reveal:   true,
							Shuffle:  true,
							Optional: true,
							Reason:   "Trophy Mage — an artifact card with mana value 3",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
