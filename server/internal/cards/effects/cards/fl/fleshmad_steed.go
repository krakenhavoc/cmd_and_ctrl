package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fleshmad Steed — Creature — Horse {1}{B}, 2/2 (Theros):
//
//	"Whenever another creature dies, tap this creature."
//
// Murder of Crows's death watcher (anyone's creature, the Steed's own death
// excluded). The Steed taps if it is still on the battlefield as the trigger
// resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5acf1424-6886-4777-be9c-a845502e08bd",
		Name:         "Fleshmad Steed",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventLTB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.B15AnotherCreatureDied(ev, source, g)
			},
			Key:     "Fleshmad Steed — tap this creature",
			Purpose: game.Purpose{DeathPayoff: true},
			Effect:  effects.OldB395TapThis,
		}},
	})
}
