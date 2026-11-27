package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tectonic Instability — Enchantment {2}{R} (Invasion):
//
//	"Whenever a land enters, tap all lands its controller controls."
//
// Zo-Zu the Punisher's trigger: every land, any controller's, however it
// entered. The controller is read off the entering land as the trigger is
// built (a Build fill-in), so a land that has left or changed hands is still
// attributed correctly; its controller's lands, the new one included, are
// tapped as the ability resolves.
//
// No simplification.
func init() {
	const label = "Tectonic Instability — tap all lands its controller controls"
	effects.Register(effects.Spec{
		OracleID:     "0e68ee36-9649-490a-8483-fc8bdc23f161",
		Name:         "Tectonic Instability",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B39LandEnteredController(ev, g)
				return ok
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, label)
				item.Params.Player, _ = effects.B39LandEnteredController(ev, g)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.B24TapAllLandsControlledBy(effects.NewContext(g, item), item.Params.Player)
			},
		}},
	})
}
