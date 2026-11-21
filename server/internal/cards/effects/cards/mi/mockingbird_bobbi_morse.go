package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mockingbird, Bobbi Morse — Legendary Creature — Human Hero {3}{W}, 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"Whenever Mockingbird is dealt damage, put a +1/+1 counter on her.
//	 (She must survive the damage to get the counter.)"
//
// The condition is Red Hulk's enrage: one trigger per batch of damage
// (CR 120.2), raised from the damage event. The counter goes on only if
// Mockingbird is still the same object on the battlefield when it
// resolves, so lethal damage (she has died by then) gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d34affa9-3629-49e1-b1e4-4742aff6eafc",
		Name:         "Mockingbird, Bobbi Morse",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventDealDamage},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return effects.B35SelfWasDealtDamage(ev, source)
			},
			Key:        "Mockingbird, Bobbi Morse — a +1/+1 counter on her",
			AtBatchEnd: effects.ThatMuchDamage(),
			Effect:     effects.PutACounterOnThis(game.CounterPlusOne),
		}},
	})
}
