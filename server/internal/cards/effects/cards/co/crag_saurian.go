package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Crag Saurian — Creature — Lizard {R}{R}{R}, 4/4 (Mercadian Masques, S82 OLD,
// #5619):
//
//	"Whenever a source deals damage to this creature, that source's controller
//	 gains control of this creature."
//
// One trigger per damage event with an amount above zero (damage that is
// prevented is not dealt). The source's controller is the damage event's actor,
// or, with none named, the controller its card last had (CR 608.2h), so damage
// from a source that has since left still counts. The control change has no
// duration (CR 611.2a) and is pinned to the Saurian as the object it is now
// (CR 400.7); a player who already controls it changes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7779e3be-5c3d-454f-88c1-9c0a97ebff09",
		Name:         "Crag Saurian",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventDealDamage, cragSaurianWasDealtDamage,
				"Crag Saurian — that source's controller gains control of this creature",
				cragSaurianChangeControl),
		},
	})
}

func cragSaurianWasDealtDamage(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
	return ev.Kind == game.EventDealDamage && ev.Amount > 0 && ev.Target == source.InstanceID
}

func cragSaurianChangeControl(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	ev := item.Trigger.Event
	taker := ev.Actor
	if taker == uuid.Nil {
		if src, ok := g.LookupCardForEffect(ev.Source); ok {
			taker = src.Controller
		}
	}
	if taker == uuid.Nil || g.PlayerByIDForEffect(taker) == nil {
		return nil
	}
	return effects.GainControl{
		Target:     item.SourceCardID,
		Controller: taker,
		Duration:   game.IndefiniteDuration(),
		Label:      "Crag Saurian — gain control",
	}.Apply(effects.NewContext(g, item))
}
