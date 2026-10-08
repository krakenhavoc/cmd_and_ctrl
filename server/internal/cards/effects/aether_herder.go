package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Herder — Creature — Elf Artificer Druid {3}{G}, 3/3:
//
//	"When this creature enters, you get {E}{E} (two energy counters).
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
		OracleID:     "be89a032-8ade-4048-b95d-76e958dca330",
		Name:         "Aether Herder",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Aether Herder", 2),
			whenThisAttacksMayPayEnergy("Aether Herder", 2, "create a 1/1 Servo", createServo),
		},
	})
}
