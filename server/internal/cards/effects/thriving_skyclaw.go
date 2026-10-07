package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thriving Skyclaw — Creature — Cat Dragon {2}{R}{R}, 3/2:
//
//	"Flying
//	 When this creature enters, you get {E}{E}{E} (three energy counters).
//	 Whenever this creature attacks, you may pay {E}{E}{E}. If you do, put a +1/+1 counter on it."
//
// ADR 0129 §3 (#1995): the energy is paid as the attack trigger resolves
// (CR 118.12), through the pay-unless prompt with an energy payment,
// which holds the declare attackers step until it is answered. "It" is
// the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d268143f-27f0-4f14-8cd3-481923aaff6d",
		Name:            "Thriving Skyclaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Thriving Skyclaw", 3),
			whenThisAttacksMayPayEnergy("Thriving Skyclaw", 3, "put a +1/+1 counter on it", thisStillHere(plusOneCountersOnThis(1))),
		},
	})
}
