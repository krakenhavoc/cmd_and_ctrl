package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starlit Mantle — Enchantment — Aura {1}{U} (Foundations Jumpstart):
//
//	"Flash
//	 Enchant creature you control
//	 When this Aura enters, enchanted creature gains hexproof until end of turn.
//	 Enchanted creature gets +1/+1."
//
// The hexproof grant is Aquitect's Defenses' shared body.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8a7643e4-b824-441a-bbe6-349fc7a3e11e",
		Name:            "Starlit Mantle",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         effects.EnchantCreature(effects.YouControl()),
		Static:          []game.StaticAbility{effects.PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Starlit Mantle — enchanted creature gains hexproof until end of turn", effects.AquitectsDefensesEnters),
		},
	})
}
