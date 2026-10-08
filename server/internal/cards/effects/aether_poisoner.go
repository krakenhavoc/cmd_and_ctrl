package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Poisoner — Creature — Human Artificer {1}{B}, 1/1:
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)
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
		OracleID:        "73757d44-7889-416b-94d2-e730e601ace3",
		Name:            "Aether Poisoner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Aether Poisoner", 2),
			whenThisAttacksMayPayEnergy("Aether Poisoner", 2, "create a 1/1 Servo", createServo),
		},
	})
}
