package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thriving Turtle — Creature — Turtle {U}, 0/3:
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
		OracleID:     "85dda096-48b2-414e-8c4d-4ffa6e5dac11",
		Name:         "Thriving Turtle",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Thriving Turtle", 2),
			whenThisAttacksMayPayEnergy("Thriving Turtle", 2, "put a +1/+1 counter on it", thisStillHere(plusOneCountersOnThis(1))),
		},
	})
}
