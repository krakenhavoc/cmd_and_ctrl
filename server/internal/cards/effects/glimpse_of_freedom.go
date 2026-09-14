package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimpse of Freedom — Instant {1}{U}:
//
//	"Draw a card.
//	 Escape—{2}{U}, Exile five other cards from your graveyard."
//
// The card that makes the point about TIMING rather than about
// value: escape does not change a spell's timing rules, so an
// instant escapes at instant speed and a sorcery does not. The whole
// difference between this and Fruit of Tizerus is which window the
// second cast can be taken in, and nothing in the engine has to know
// that — the sorcery-speed gate reads the card's type line exactly as
// it does for a cast from hand.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "02d78e76-8151-4851-bb4e-e0a1fa98756f",
		Name:             "Glimpse of Freedom",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Escape("{2}{U}", 5)},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return DrawCards{Player: item.Controller, N: 1}.Apply(ctx)
		},
	})
}
