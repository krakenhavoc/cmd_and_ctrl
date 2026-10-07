package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thriving Grubs — Creature — Gremlin {1}{R}, 2/1:
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
		OracleID:     "30ec09e4-82bf-4a6b-b6c8-ff267447e0a9",
		Name:         "Thriving Grubs",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Thriving Grubs", 2),
			whenThisAttacksMayPayEnergy("Thriving Grubs", 2, "put a +1/+1 counter on it", thisStillHere(plusOneCountersOnThis(1))),
		},
	})
}
