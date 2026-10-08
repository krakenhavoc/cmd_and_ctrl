package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Inspector — Creature — Dwarf Artificer {3}{W}, 2/3:
//
//	"Vigilance
//	 When this creature enters, you get {E}{E} (two energy counters).
//	 Whenever this creature attacks, you may pay {E}{E}. If you do, create a 1/1 colorless Servo artifact creature token."
//
// ADR 0129 §3 (#1995): the energy is paid as the attack trigger resolves
// (CR 118.12), through the pay-unless prompt with an energy payment,
// which holds the declare attackers step until it is answered. "It" is
// the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9373a176-a5e9-4fdc-906b-bc6aef657e5b",
		Name:            "Aether Inspector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Aether Inspector", 2),
			whenThisAttacksMayPayEnergy("Aether Inspector", 2, "create a 1/1 Servo", createServo),
		},
	})
}
