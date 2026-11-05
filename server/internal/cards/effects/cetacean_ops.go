package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cetacean Ops — Creature — Whale Officer {3}{U}, 3/3 (Star Trek):
//
//	Ward {1}
//	Whenever you cast a noncreature spell, put a +1/+1 counter on this
//	creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "64e39ec9-358a-4aa0-a7e9-164e8aa0b236",
		Name:         "Cetacean Ops",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Ward(WardMana("{1}"), "Cetacean Ops — ward {1}"),
			WheneverYouCast(Noncreature(), "Cetacean Ops — put a +1/+1 counter on this creature",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
	})
}
