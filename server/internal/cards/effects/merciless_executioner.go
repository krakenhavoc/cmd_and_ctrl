package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Merciless Executioner — Creature — Orc Warrior {2}{B}, 3/1 (EDHREC
// rank 2205):
//
//	"When this creature enters, each player sacrifices a creature of
//	 their choice."
//
// Fleshbag Marauder with a different type line, and the same body:
// EachPlayerSacrifices with the controller included — "each player"
// means you too, and the Executioner is on the battlefield when its
// own trigger resolves, so feeding it to itself is the printed line.
// Every player picks their own creature in their own prompt; a
// player with none is skipped (CR 701.17b).
//
// The ETB is a triggered ability and uses the stack (#578). It used to
// run from the direct AsEnters hook, which gave nobody a response
// window; now the trigger waits for every player to pass, like every
// other "When ~ enters".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c3c45d50-9038-41df-bb2f-9bc40071845b",
		Name:         "Merciless Executioner",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.CardID == source.InstanceID
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Merciless Executioner — each player sacrifices a creature",
					func(g *game.Game, item *game.StackItem) error {
						return EachPlayerSacrifices{
							Match: Creature(),
							Label: "a creature",
						}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
