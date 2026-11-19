package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "8a7643e4-b824-441a-bbe6-349fc7a3e11e",
		Name:            "Starlit Mantle",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Targets:         EnchantCreature(YouControl()),
		Static:          []game.StaticAbility{PumpAttached(1, 1)},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Starlit Mantle — enchanted creature gains hexproof until end of turn", aquitectsDefensesEnters),
		},
	})
}
