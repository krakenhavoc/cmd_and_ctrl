package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Vile Entomber — Creature — Zombie Warlock {2}{B}{B}, 2/2 (EDHREC
// rank 1906):
//
//	"Deathtouch (Any amount of damage this deals to a creature is
//	 enough to destroy it.)
//	 When this creature enters, search your library for a card, put
//	 that card into your graveyard, then shuffle."
//
// Entomb on a body. The ETB is a mandatory search for any card to
// the graveyard; the S22 searcher chooses, and the library shuffles
// afterwards. The search is not a reveal — the card goes to a
// public zone, which is what shows it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "556dd333-4066-41f7-98f6-41794754de71",
		Name:            "Vile Entomber",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: b06SelfETB,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Vile Entomber — search for a card, put it into your graveyard",
					func(g *game.Game, item *game.StackItem) error {
						return SearchLibrary{
							Player:    item.Controller,
							Predicate: func(game.Card) bool { return true },
							Dest:      game.ZoneGraveyard,
							Limit:     1,
							Shuffle:   true,
							Reason:    "Vile Entomber — a card, into your graveyard",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
