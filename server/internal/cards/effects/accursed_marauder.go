package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Accursed Marauder — Creature — Zombie Warrior {1}{B}, 2/1 (EDHREC
// rank 476):
//
//	"When this creature enters, each player sacrifices a nontoken
//	 creature of their choice."
//
// Fleshbag Marauder for two mana with one word added: NONTOKEN. A
// Treasure deck cannot feed it a Zombie token, and neither can you —
// but the Marauder itself is nontoken and on the battlefield when
// its trigger resolves, so eating itself is still the free line.
// EachPlayerSacrifices fans out one prompt per player, each offering
// only that player's own nontoken creatures.
//
// The ETB is a triggered ability and uses the stack (#578). It used to
// run from the direct AsEnters hook, which gave nobody a response
// window; now the trigger waits for every player to pass, like every
// other "When ~ enters".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8ad23a1-0b43-48ea-9fbe-d89b29194509",
		Name:         "Accursed Marauder",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Accursed Marauder — each player sacrifices a nontoken creature",
					func(g *game.Game, item *game.StackItem) error {
						return EachPlayerSacrifices{
							Match: b04NontokenCreature,
							Label: "a nontoken creature",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
