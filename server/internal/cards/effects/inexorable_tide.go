package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Inexorable Tide — Enchantment {3}{U}{U}:
//
//	"Whenever you cast a spell, proliferate."
//
// Flux Channeler with no creature clause and no body: EVERY spell,
// creatures included, and nothing can attack it down. Five mana that
// does nothing the turn it lands and wins the game two turns later
// in a deck built around it.
//
// It triggers on its own cast? No — the Tide is not on the
// battlefield while it is being cast, so there is nothing to
// trigger. The first spell it sees is the next one.
func init() {
	Register(Spec{
		OracleID:     "40ec0a47-badf-4074-b0a8-749bb7c17b95",
		Name:         "Inexorable Tide",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You don't choose what to proliferate — the game picks for you, adding every counter that helps you and every counter that hurts an opponent."},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Inexorable Tide — proliferate",
					func(g *game.Game, item *game.StackItem) error {
						return Proliferate{}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
