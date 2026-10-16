package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// All Will Be One — Enchantment {3}{R}{R} (EDHREC rank 1401):
//
//	"Whenever you put one or more counters on a permanent or player,
//	 this enchantment deals that much damage to target opponent,
//	 creature an opponent controls, or planeswalker an opponent
//	 controls."
//
// The counters deck's damage engine: every +1/+1 counter is a ping
// and a Cathars' Crusade trigger is a Pyroclasm. The condition reads
// the counter-placement event and works out, from the event log,
// how many counters landed (the event carries only the new total), and
// takes who put them there off the event — see b12CountersPlacedByYou.
// One event is one permanent, so
// "one or more counters on a permanent" is one trigger per permanent
// touched, which is the printed batching. The target clause is the
// printed one: an opponent, or a creature or planeswalker one
// controls.
//
// "Or player" (#3172): counters put on a player (poison, energy, rad)
// emit EventPlayerCounterPlaced, whose Amount is the delta that landed
// and whose Actor is the placer, which every placement names since
// #3942. A removal (negative delta) is not a placement.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "477374dc-042c-48f7-9ebe-99c15d8ae04f",
		Name:         "All Will Be One",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventCounterPlaced, game.EventPlayerCounterPlaced},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := awboCountersPlacedByYou(ev, source, g)
				return ok
			},
			Targets: b12TargetOpponentOrTheirCreatureOrPlaneswalker(),
			Key:     "All Will Be One — that much damage",
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				amount, _ := awboCountersPlacedByYou(ev, source, g)
				item := game.NewTriggeredItem(source, "All Will Be One — that much damage")
				item.Params.Amount = amount
				return item
			},
			Effect: DealAmountFromParamsToFirstTarget,
		}},
	})
}

// awboCountersPlacedByYou is All Will Be One's condition: counters put
// on a permanent (b12CountersPlacedByYou) or on a player. Both halves
// read the placer off the event.
func awboCountersPlacedByYou(ev game.Event, source *game.Card, g *game.Game) (int, bool) {
	if ev.Kind != game.EventPlayerCounterPlaced {
		return b12CountersPlacedByYou(ev, source, g)
	}
	if ev.Amount <= 0 || ev.Actor == uuid.Nil || ev.Actor != source.Controller {
		return 0, false
	}
	return ev.Amount, true
}
