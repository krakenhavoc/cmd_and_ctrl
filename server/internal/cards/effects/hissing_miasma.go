package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hissing Miasma — Enchantment {1}{B}{B}:
//
//	"Whenever a creature attacks you, its controller loses 1 life."
//
// Revenge of Ravens without the life gain and without the
// planeswalker clause. "Attacks YOU" is the player itself: an
// EventAttack names the attacked player or planeswalker in Target, so
// the trigger passes only when that is the enchantment's controller —
// a creature attacking one of your planeswalkers does not trigger it.
// One trigger per attacking creature, with the attacker's
// controller carried into the item by value so the loss lands on the
// right player even after the attacker has left.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e257d8e0-06e9-433d-a750-1962db399388",
		Name:         "Hissing Miasma",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventAttack},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller && ev.Target == source.Controller
			},
			Key: "Hissing Miasma — the attacker's controller loses 1 life",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, "Hissing Miasma — the attacker's controller loses 1 life")
				item.Params.Player = ev.Actor
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				victim := item.Params.Player
				if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, victim, -1)
			},
		}},
	})
}
