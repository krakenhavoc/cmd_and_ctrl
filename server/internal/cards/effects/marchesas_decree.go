package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Marchesa's Decree — Enchantment {3}{B} (Foundations Commander):
//
//	"When this enchantment enters, you become the monarch.
//	 Whenever a creature attacks you or a planeswalker you control, that
//	 creature's controller loses 1 life."
//
// The enters trigger is the monarch step; the second ability is Revenge of
// Ravens without the life gain: one trigger per attacking creature aimed at
// the controller or a planeswalker they control (CR 508.3a; an attack on a
// battle does not trigger, #4864). The attacker's controller is carried
// into the item by value so the loss lands on the right player after the
// attacker has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e4ab5ee4-4683-4219-9e15-31975e69efbb",
		Name:         "Marchesa's Decree",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Marchesa's Decree — you become the monarch", Do(BecomeTheMonarch{})),
			{
				Watches: []game.EventKind{game.EventAttack},
				AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
					return ev.Kind == game.EventAttack && ev.Actor != uuid.Nil &&
						attacksYouOrYourPlaneswalker(g, ev.Target, source.Controller)
				},
				Key: "Marchesa's Decree — the attacker's controller loses 1 life",
				Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
					item := game.NewTriggeredItem(source, "Marchesa's Decree — the attacker's controller loses 1 life")
					item.Params.Player = ev.Actor
					return item
				},
				Effect: cmdrB3AttackersControllerLosesOne,
			},
		},
	})
}
