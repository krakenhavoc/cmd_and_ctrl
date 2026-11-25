package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Glaring Aegis — Enchantment — Aura {W} (Core Set 2020, #5518):
//
//	"Enchant creature
//	 When this Aura enters, tap target creature an opponent controls.
//	 Enchanted creature gets +1/+3."
//
// Cartouche of Zeal's shape: a layer 7c pump on the host and an enters trigger
// whose target (an opponent's creature, not necessarily the enchanted one) is
// chosen as the trigger goes on the stack and tapped if still legal as it
// resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0b0c98ba-3505-40aa-b0f2-e03c9790f51e",
		Name:         "Glaring Aegis",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(1, 3)},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Glaring Aegis — tap target creature an opponent controls", effects.TapEachLegalTarget),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
