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
// both how many counters landed (the event carries only the new
// total) and whether the source's controller is the one who put them
// there (the player whose spell or ability was resolving) — see
// b12CountersPlacedByYou for the attribution rules and why every one
// of them errs toward NOT firing. One event is one permanent, so
// "one or more counters on a permanent" is one trigger per permanent
// touched, which is the printed batching. The target clause is the
// printed one: an opponent, or a creature or planeswalker one
// controls.
//
// "Or player" (#3172): counters put on a player (poison, energy, rad)
// emit EventPlayerCounterPlaced, whose Amount is the delta that landed
// and whose Actor is the placer when the placement named one. When it
// did not, the player resolving the item is credited, exactly as for a
// permanent. A removal (negative delta) is not a placement.
//
// Sandbox simplification, declared:
//
//   - Counters whose placer the log cannot attribute — a loyalty
//     cost or a Saga's lore counter on your own permanent while an
//     opponent's spell was the last thing to resolve — do not fire
//     it. Weaker, never stronger.
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
// on a permanent (b12CountersPlacedByYou) or on a player. The player
// half takes the placer off the event and falls back to the resolving
// controller when the placement named nobody — weaker, never stronger,
// where the log cannot say.
func awboCountersPlacedByYou(ev game.Event, source *game.Card, g *game.Game) (int, bool) {
	if ev.Kind != game.EventPlayerCounterPlaced {
		return b12CountersPlacedByYou(ev, source, g)
	}
	if ev.Amount <= 0 {
		return 0, false
	}
	placer := ev.Actor
	if placer == uuid.Nil {
		placer = b12ResolvingController(ev, g)
	}
	if placer != source.Controller {
		return 0, false
	}
	return ev.Amount, true
}
