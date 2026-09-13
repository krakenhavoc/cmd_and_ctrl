package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fate Unraveler — Enchantment Creature — Hag {3}{B}, 3/4 (EDHREC
// rank 1707):
//
//	"Whenever an opponent draws a card, this creature deals 1 damage
//	 to that player."
//
// Underworld Dreams on a body — and an enchantment body, so it is
// both a creature and an enchantment for anything that counts
// either. Razorkin Needlehead's trigger exactly: EventDrawCard fires
// once per card, the drawer's ID is captured in Build by value (the
// Edric pattern), and the Hag is the damage source, so a Fog-class
// shield stops it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d66f67d-3148-4636-ac5a-2cf8a51d5e50",
		Name:         "Fate Unraveler",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				drawer := ev.Actor
				return game.NewTriggeredItem(source, "Fate Unraveler — 1 damage to the player who drew",
					func(g *game.Game, item *game.StackItem) error {
						if g.PlayerByIDForEffect(drawer) == nil {
							return nil
						}
						return DealDamage{Source: item.SourceCardID, Target: drawer, Amount: 1}.Apply(NewContext(g, item))
					})
			},
		}},
	})
}
