package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Riparian Tiger — Creature — Cat {3}{G}{G}, 4/4:
//
//	"Trample
//	 When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}{E}. If you do, it gets +2/+2 until end of turn."
//
// ADR 0129 §3 (#1995): the energy is paid as the attack trigger resolves
// (CR 118.12), through the pay-unless prompt with an energy payment,
// which holds the declare attackers step until it is answered. "It" is
// the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "dbb6b7f0-33fd-4530-a621-114148fb3929",
		Name:            "Riparian Tiger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Riparian Tiger", 2),
			whenThisAttacksMayPayEnergy("Riparian Tiger", 2, "have it get +2/+2", thisGetsUntilEndOfTurn(2, 2, "Riparian Tiger — +2/+2 until end of turn")),
		},
	})
}
