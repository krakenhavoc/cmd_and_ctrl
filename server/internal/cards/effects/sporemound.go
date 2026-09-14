package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sporemound — Creature — Fungus {3}{G}{G}, 3/3 (EDHREC rank 3591):
//
//	"Landfall — Whenever a land you control enters, create a 1/1
//	 green Saproling creature token."
//
// The landfall Saproling engine. Fires once per land, however it
// entered — played, fetched, or returned from a graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1be56a3d-a6c0-4b65-ae71-3d90ceefc6c0",
		Name:         "Sporemound",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b33LandYouControlEntered(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Sporemound — create a 1/1 green Saproling", b34CreateTokens(b11GreenSaprolingToken, 1))
			},
		}},
	})
}
