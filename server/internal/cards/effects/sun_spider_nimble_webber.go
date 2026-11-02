package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sun-Spider, Nimble Webber — Legendary Creature — Spider Human Hero
// {3}{W/U}, 3/2 (Marvel's Spider-Man, #4460):
//
//	"During your turn, Sun-Spider has flying.
//	 When Sun-Spider enters, search your library for an Aura or Equipment
//	 card, reveal it, put it into your hand, then shuffle."
//
// The flying is Cid, Freeflier Pilot's Jump: a layer 6 grant on the creature
// itself that is on only while it is its controller's turn, re-read every
// recompute. The search is mandatory and reveals the card; the library is
// shuffled even when nothing is found.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d961997-4f5f-4ce2-b062-11411147e2f7",
		Name:         "Sun-Spider, Nimble Webber",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			KeywordGrant(func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && target.IsCreature() && isActivePlayer(g, source.Controller)
			}, "flying"),
		},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Sun-Spider, Nimble Webber — search for an Aura or Equipment card, reveal it, put it into your hand, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return c.HasSubtype("Aura") || c.HasSubtype("Equipment") },
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Sun-Spider, Nimble Webber — an Aura or Equipment card, revealed and put into your hand",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
