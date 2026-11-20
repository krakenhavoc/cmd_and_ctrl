package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Blood Reckoning — Enchantment {3}{B} (Magic 2013, #5568):
//
//	"Whenever a creature attacks you or a planeswalker you control, that
//	 creature's controller loses 1 life."
//
// Marchesa's Decree's attack trigger alone: one trigger per creature
// attacking the controller or a planeswalker they control
// (attacksYouOrYourPlaneswalker, CR 508.3a). The attacker's controller is
// carried into the item by value, so the loss lands on the right player
// even after the attacker has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bd5310df-2e6c-4584-9a5a-03d6d06b3c20",
		Name:         "Blood Reckoning",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil &&
					attacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
			},
			Key: "Blood Reckoning — the attacker's controller loses 1 life",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Blood Reckoning — the attacker's controller loses 1 life")
				item.Params.Player = ev.Actor
				return item
			},
			Effect: cmdrB3AttackersControllerLosesOne,
		}},
	})
}
