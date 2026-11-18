package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Horobi, Death's Wail — Legendary Creature — Spirit {2}{B}{B}, 4/4:
//
//	"Flying
//	 Whenever a creature becomes the target of a spell or ability,
//	 destroy that creature."
//
// Any creature, whoever controls it or aims at it, Horobi included. The
// trigger goes on the stack above the spell or ability, so it resolves
// first; the creature is the one the event names, and one that has left
// the battlefield by then is nothing to destroy.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "44e8d51e-0885-49df-a77c-cf4a88e41bc1",
		Name:            "Horobi, Death's Wail",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventBecomesTarget},
			Key:     "Horobi, Death's Wail — destroy that creature",
			AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventBecomesTarget {
					return false
				}
				if z := g.FindCardZoneForEffect(ev.CardID); z == nil || z.Kind != game.ZoneBattlefield {
					return false // a creature spell on the stack is not a creature
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				id := item.Trigger.Event.CardID
				if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneBattlefield {
					return nil
				}
				return DestroyTarget{Target: id}.Apply(NewContext(g, item))
			},
		}},
	})
}
