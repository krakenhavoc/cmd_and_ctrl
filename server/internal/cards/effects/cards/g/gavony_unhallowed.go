package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gavony Unhallowed — Creature — Zombie {3}{B}, 2/4 (Eldritch Moon, #5541):
//
//	"Whenever another creature you control dies, put a +1/+1 counter on
//	 this creature."
//
// Unruly Mob's trigger, word for word.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "95ccb4a6-6e45-4c00-b7c6-25384022c02f",
		Name:         "Gavony Unhallowed",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.AllOf(effects.AnotherCreatureDied, effects.ACreatureYouControlDied),
				"Gavony Unhallowed — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
