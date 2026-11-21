package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Thriving Rhino — Creature — Rhino {2}{G}, 2/3:
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
	effects.Register(effects.Spec{
		OracleID:     "d8841f3a-f3ff-42ca-89a8-0cc1c3ca6a6c",
		Name:         "Thriving Rhino",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 2},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Thriving Rhino", 2),
			effects.WhenThisAttacksMayPayEnergy("Thriving Rhino", 2, "put a +1/+1 counter on it", effects.EnergyPlusOneCounterOnThis),
		},
	})
}
