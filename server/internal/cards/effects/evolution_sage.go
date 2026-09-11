package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Evolution Sage — Creature — Elf Druid {2}{G}, 3/2:
//
//	"Landfall — Whenever a land you control enters, proliferate."
//
// Lotus Cobra's trigger with a different payload: every land drop,
// every fetchland crack, every Cultivate is a free proliferate. The
// filter is the shared landfall helper, so "a land YOU control"
// excludes an opponent's land drop and includes a land that entered
// under your control from anywhere — played, fetched or reanimated.
func init() {
	Register(Spec{
		OracleID: "b45fdeab-00cc-4422-af9f-66f30a880a7c",
		Name:     "Evolution Sage",
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := enteredUnderYourControl(ev, source, g, false)
				return ok && c.IsLand()
			},
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				return game.NewTriggeredItem(source, "Evolution Sage — proliferate (landfall)",
					func(g *game.Game, item *game.StackItem) error {
						return Proliferate{}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
