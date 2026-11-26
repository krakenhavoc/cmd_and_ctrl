package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Reaper of Sheoldred — Creature — Phyrexian Horror {4}{B}, 2/5 (New Phyrexia,
// #5575):
//
//	"Infect
//	 Whenever a source deals damage to this creature, that source's controller
//	 gets a poison counter."
//
// One trigger per damage event dealt to the Reaper (CR 603.2c), whoever's source
// and whether combat or not, and "that source's controller" is the damage
// event's actor (or, with none recorded, the source card's controller), the
// controller of the source as it dealt the damage (CR 608.2h).
// Damage that was prevented is not dealt, so it triggers nothing. The poison
// counter is put there by the Reaper's controller (CR 122.1f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "251cee41-30f5-4bba-93cd-31493a5ea051",
		Name:            "Reaper of Sheoldred",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"infect"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source) && damageSourceController(ev, g) != uuid.Nil
			}, "Reaper of Sheoldred — that source's controller gets a poison counter",
				func(g *game.Game, item *game.StackItem) error {
					ev := item.Trigger.Event
					return g.AddPlayerCounterByForEffect(item.Controller, damageSourceController(ev, g), game.CounterPoison, 1)
				}),
		},
	})
}

// damageSourceController is the controller of the source of a damage event as it
// dealt the damage: the event's actor, or, for damage the engine recorded without
// one, the controller of the source card as it last stood (CR 608.2h).
func damageSourceController(ev game.Event, g *game.Game) uuid.UUID {
	if ev.Actor != uuid.Nil {
		return ev.Actor
	}
	if c, ok := g.LookupCardForEffect(ev.Source); ok {
		return c.Controller
	}
	return uuid.Nil
}
