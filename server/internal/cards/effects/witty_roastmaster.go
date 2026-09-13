package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Witty Roastmaster — Creature — Devil Citizen {2}{R}, 3/2 (EDHREC
// rank 1471):
//
//	"Alliance — Whenever another creature you control enters, this
//	 creature deals 1 damage to each opponent."
//
// Impact Tremors on a body. One trigger per creature entering, and
// the damage is the Roastmaster's, so a doubler doubles it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "34edbf1b-0826-49f1-a04d-15cc8c543b22",
		Name:         "Witty Roastmaster",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b13AnotherCreatureYouControlEntered(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Witty Roastmaster — 1 damage to each opponent",
					func(g *game.Game, item *game.StackItem) error {
						return damageToEachOpponent(g, item, 1)
					})
			},
		}},
	})
}
