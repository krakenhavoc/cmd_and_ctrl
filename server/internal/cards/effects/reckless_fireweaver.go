package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Reckless Fireweaver — 1/3 Creature — Human Artificer for {1}{R}:
//
//	"Whenever an artifact you control enters, this creature deals 1
//	damage to each opponent."
//
// The reason an Izzet artifact deck can win without attacking: every
// Treasure, Clue and Blood token that hits the battlefield pings the
// whole table. See artifactEnteredUnderYourControl for the
// one-trigger-per-artifact batching note.
func init() {
	Register(Spec{
		OracleID: "180e1a7e-890d-477c-80a5-da8a5f2857b3",
		Name:     "Reckless Fireweaver",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return artifactEnteredUnderYourControl(ev, source, g)
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Reckless Fireweaver — 1 damage to each opponent",
					func(g *game.Game, item *game.StackItem) error {
						return damageToEachOpponent(g, item, 1)
					})
			},
		}},
	})
}
