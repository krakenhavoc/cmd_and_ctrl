package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bumi, Eclectic Earthbender — Legendary Creature — Human Noble Ally
// {3}{G}{G}, 4/4 (Avatar: The Last Airbender Eternal, #5075):
//
//	"When Bumi enters, earthbend 1. (Target land you control becomes a
//	 0/0 creature with haste that's still a land. Put a +1/+1 counter on
//	 it. When it dies or is exiled, return it to the battlefield tapped.)
//	 Whenever Bumi attacks, put two +1/+1 counters on each land creature
//	 you control."
//
// The entry is the earthbend keyword action with its own target clause
// (Badgermole Cub's shape). The attack trigger counts the land creatures
// you control as it resolves, through the layers, so a land earthbent
// earlier (this turn's or another card's) gets its two counters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "301af2b2-e2f7-4220-a3f5-7702fb62664a",
		Name:         "Bumi, Eclectic Earthbender",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			avB14WhenEntersEarthbend("Bumi, Eclectic Earthbender — earthbend 1", 1),
			WheneverThisAttacks("Bumi, Eclectic Earthbender — two +1/+1 counters on each land creature you control",
				avB38TwoCountersOnEachLandCreatureYouControl),
		},
	})
}
