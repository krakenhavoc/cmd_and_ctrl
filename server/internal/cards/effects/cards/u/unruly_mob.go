package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unruly Mob — Creature — Human {1}{W}, 1/1 (Innistrad: Midnight Hunt
// Commander reprint, S82):
//
//	"Whenever another creature you control dies, put a +1/+1 counter on
//	 this creature."
//
// Feral Ghoul's dies trigger: AnotherCreatureDied with ACreatureYouControlDied,
// the controller read as the creature last existed (CR 603.10a). The Mob
// itself dying is not "another".
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05d0b904-a680-4f2f-aaa1-e227a7fe7e79",
		Name:         "Unruly Mob",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.AllOf(effects.AnotherCreatureDied, effects.ACreatureYouControlDied),
				"Unruly Mob — put a +1/+1 counter on it", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
