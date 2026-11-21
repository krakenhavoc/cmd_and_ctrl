package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cetacean Ops — Creature — Whale Officer {3}{U}, 3/3 (Star Trek):
//
//	Ward {1}
//	Whenever you cast a noncreature spell, put a +1/+1 counter on this
//	creature.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64e39ec9-358a-4aa0-a7e9-164e8aa0b236",
		Name:         "Cetacean Ops",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Ward(effects.WardMana("{1}"), "Cetacean Ops — ward {1}"),
			effects.WheneverYouCast(effects.Noncreature(), "Cetacean Ops — put a +1/+1 counter on this creature",
				effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
