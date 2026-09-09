package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Weftstalker Ardent — 2/3 Creature — Drix Artificer for {2}{R}:
//
//	"Whenever another creature or artifact you control enters, this
//	 creature deals 1 damage to each opponent.
//	 Warp {R}"
//
// Reckless Fireweaver widened to creatures as well as artifacts,
// which in this deck means every Treasure AND every Pirate pings the
// table. The "another" is load-bearing: it doesn't ping off its own
// arrival.
//
// Warp is an alternative cast path (S29) and isn't modelled.
func init() {
	Register(Spec{
		OracleID: "926d52a5-4db1-46ce-9567-17c28bf56ae7",
		Name:     "Weftstalker Ardent",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, true)
				return ok && (c.IsCreature() || c.IsArtifact())
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Weftstalker Ardent — 1 damage to each opponent",
					func(g *game.Game, item *game.StackItem) error {
						return damageToEachOpponent(g, item, 1)
					})
			},
		}},
	})
}
