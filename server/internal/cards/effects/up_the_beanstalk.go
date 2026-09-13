package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Up the Beanstalk — Enchantment {1}{G} (EDHREC rank 1111):
//
//	"When this enchantment enters and whenever you cast a spell with
//	 mana value 5 or greater, draw a card."
//
// One printed ability with two trigger conditions, so one
// TriggeredAbility watching two event kinds (the Sun Titan shape):
// its own entry, or the controller casting a spell whose mana value
// ON THE STACK is five or more. That is CR 202.3e — X counts at the
// announced value, so a Fireball for X=4 is a five-drop here, exactly
// as it is in paper. The cast trigger goes on the stack above the
// spell and resolves first.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "050f5733-7c0b-4991-9a6c-7ea12ccf0ca9",
		Name:         "Up the Beanstalk",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB, game.EventCast},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind == game.EventETB {
					return ev.CardID == source.InstanceID
				}
				return b09SpellCastByYouWithManaValueAtLeast(ev, source, g, 5)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Up the Beanstalk — draw a card",
					func(g *game.Game, item *game.StackItem) error {
						return DrawCards{Player: item.Controller, N: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
