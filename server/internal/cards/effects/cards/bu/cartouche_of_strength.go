package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cartouche of Strength — Enchantment — Aura Cartouche {2}{G} (Amonkhet, S82 OLD batch 244, #5535):
//
//	"Enchant creature you control
//	 When this Aura enters, you may have enchanted creature fight target
//	 creature an opponent controls. (Each deals damage equal to its power to
//	 the other.)
//	 Enchanted creature gets +1/+1 and has trample."
//
// Pitiless Fists' enters fight with a "you may": the target is chosen as the
// trigger goes on the stack, the question is asked as it resolves, and the
// fight reads the enchanted creature as it then is (CR 701.14).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cba6563e-96e5-4cd3-bddf-0cf1368f9e15",
		Name:         "Cartouche of Strength",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.YouControl()),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("trample"),
		},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.Optional(
					effects.WhenThisEnters("Cartouche of Strength — you may have enchanted creature fight target creature an opponent controls",
						effects.EnchantedCreatureFightsTargetCreature),
					"Cartouche of Strength — have enchanted creature fight target creature an opponent controls?"),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
