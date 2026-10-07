package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Swooper — Creature — Vedalken Artificer {1}{U}, 1/2:
//
//	"Flying
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
		OracleID:        "f7ed301a-dce7-49d0-a68a-d3a099201f65",
		Name:            "Aether Swooper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Aether Swooper", 2),
			whenThisAttacksMayPayEnergy("Aether Swooper", 2, "create a 1/1 Servo", createServo),
		},
	})
}
