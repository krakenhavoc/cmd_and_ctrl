package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Razorkin Needlehead — Creature — Human Assassin, {R}{R}, 2/2
// (EDHREC rank 970):
//
//	"This creature has first strike during your turn.
//	 Whenever an opponent draws a card, this creature deals 1 damage
//	 to them."
//
// The red Sheoldred: a point of damage per opposing draw, from a
// two-drop. Two abilities. "First strike during your turn" is
// Zurgo Helmsmasher's conditional static — a layer 6 grant whose
// predicate reads the turn, re-answered on every recompute, and the
// layer engine invalidates on turn change so the keyword really does
// blink off on an opponent's turn. The draw trigger is Sheoldred's,
// with damage in place of life loss: Needlehead is the source, so the
// point is red damage (Torbran adds to it, a Fog-class shield stops
// it) and the drawer's ID is captured in Build by value.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a78f981a-bf8a-42a4-b171-d655cc2cc1a2",
		Name:         "Razorkin Needlehead",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Layer: game.Layer6Ability,
			AppliesTo: func(target *game.Card, g *game.Game, source *game.Card) bool {
				return target.InstanceID == source.InstanceID && isActivePlayer(g, source.Controller)
			},
			Apply: func(c *game.Characteristic, _ *game.Card, _ *game.Game, _ *game.Card) {
				for _, a := range c.Abilities {
					if a == "first strike" {
						return
					}
				}
				c.Abilities = append(c.Abilities, "first strike")
			},
		}},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDrawCard},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor != uuid.Nil && ev.Actor != source.Controller
			},
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				drawer := ev.Actor
				return game.NewTriggeredItem(source, "Razorkin Needlehead — 1 damage to the player who drew",
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
