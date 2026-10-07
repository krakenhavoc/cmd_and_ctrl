package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thriving Rats — Creature — Rat {1}{B}, 1/2:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}{E}. If you do, put a +1/+1 counter on it."
//
// ADR 0129 §3 (#1995): the energy is paid as the attack trigger resolves
// (CR 118.12), through the pay-unless prompt with an energy payment,
// which holds the declare attackers step until it is answered. "It" is
// the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0bb8500a-bc29-43f1-ab22-2b8adeb9c99d",
		Name:         "Thriving Rats",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Thriving Rats", 2),
			whenThisAttacksMayPayEnergy("Thriving Rats", 2, "put a +1/+1 counter on it", thisStillHere(plusOneCountersOnThis(1))),
		},
	})
}
