package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Aether Chaser — Creature — Human Artificer {1}{R}, 2/1:
//
//	"First strike
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
	effects.Register(effects.Spec{
		OracleID:        "8aab0b55-2779-43c7-a3ee-151b8e5c72f3",
		Name:            "Aether Chaser",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Purpose:         game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Aether Chaser", 2),
			effects.WhenThisAttacksMayPayEnergy("Aether Chaser", 2, "create a 1/1 Servo", effects.EnergyCreateServo),
		},
	})
}
