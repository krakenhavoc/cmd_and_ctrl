package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Goblin Matron — Creature — Goblin {2}{R}, 1/1 (EDHREC rank 1787):
//
//	"When this creature enters, you may search your library for a
//	 Goblin card, reveal that card, put it into your hand, then
//	 shuffle."
//
// The Goblin deck's tutor on legs. Recruiter of the Guard's shape
// with a subtype for the toughness clause: the tutor is optional, so
// the prompt always opens and "fail to find" is always an answer.
// Effective subtypes on a library card are its printed ones.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "145737a7-c597-4dec-b752-207c2d0501e3",
		Name:         "Goblin Matron",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Goblin Matron — search for a Goblin card",
					func(g *game.Game, item *game.StackItem) error {
						return SearchLibrary{
							Player:    item.Controller,
							Predicate: func(c game.Card) bool { return c.HasSubtype("Goblin") },
							Dest:      game.ZoneHand,
							Limit:     1,
							Reveal:    true,
							Shuffle:   true,
							Optional:  true,
							Reason:    "Goblin Matron — a Goblin card, revealed, to hand",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
