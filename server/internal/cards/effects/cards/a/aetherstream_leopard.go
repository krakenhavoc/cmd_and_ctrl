package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

var aetherstreamLeopardPumpPaid = effects.EnergyPaid("pay-unless/aetherstream-leopard-pump",
	effects.ThisGetsUntilEndOfTurn(2, 0, "Aetherstream Leopard — +2/+0 until end of turn"))

// Aetherstream Leopard — Creature — Cat {2}{G}, 2/3:
//
//	"Trample
//	 When this creature enters, you get {E} (an energy counter).
//	 Whenever this creature attacks, you may pay {E}. If you do, it gets +2/+0 until end of turn."
//
// ADR 0129 §3 (#1995): the energy is paid as the attack trigger resolves
// (CR 118.12), through the pay-unless prompt with an energy payment,
// which holds the declare attackers step until it is answered. "It" is
// the creature that attacked (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8517e5f4-89c5-4879-9b4f-b35e268798df",
		Name:            "Aetherstream Leopard",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Energy: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEntersYouGetEnergy("Aetherstream Leopard", 1),
			effects.WhenThisAttacksMayPayEnergy("Aetherstream Leopard", 1, "have it get +2/+0", aetherstreamLeopardPumpPaid),
		},
	})
}
