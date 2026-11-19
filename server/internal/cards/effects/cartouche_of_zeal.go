package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Cartouche of Zeal — {R} Enchantment — Aura Cartouche: "Enchant
// creature you control. When this Aura enters, target creature can't
// block this turn. Enchanted creature gets +1/+1 and has haste."
func init() {
	const label = "Cartouche of Zeal — target creature can't block this turn"
	Register(Spec{
		OracleID:     "4fa23be3-a192-45ab-bdf6-646d4be305e1",
		Name:         "Cartouche of Zeal",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("haste"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters(label, targetCreatureCantBlockThisTurn(label)),
				TargetCreature("target creature")),
		},
	})
}
