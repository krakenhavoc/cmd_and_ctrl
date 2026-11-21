package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gathering Throng — Creature — Human Citizen {2}{W}, 3/1 (Streets of New
// Capenna, #5491):
//
//	"When this creature enters, you may search your library for any number of
//	 cards named Gathering Throng, reveal them, put them into your hand, then
//	 shuffle.
//	 A deck can have any number of cards named Gathering Throng."
//
// Andromedan Automatons' optional unbounded search, matching by printed name
// (game.HasName). Declining searches nothing and does not shuffle. The
// deck-construction line needs nothing here: the deck validator reads "A deck
// can have any number of cards named …" from the oracle text (CR 113.6n).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "bbd340ce-cf06-43df-a060-be80bc46e985",
		Name:         "Gathering Throng",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Gathering Throng — search for any number of cards named Gathering Throng",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Predicate: func(c game.Card) bool { return game.HasName(c, "Gathering Throng") },
						Dest:      game.ZoneHand,
						Unbounded: true,
						Reveal:    true,
						Shuffle:   true,
						Optional:  true,
						Reason:    "Gathering Throng — you may search for any number of cards named Gathering Throng",
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
