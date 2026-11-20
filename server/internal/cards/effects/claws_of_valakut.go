package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Claws of Valakut — Enchantment — Aura {1}{R}{R} (Worldwake, #5582):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+0 for each Mountain you control and has
//	 first strike."
//
// Granite Grip's per-Mountain pump, counted for the Aura's controller,
// with Ethereal Armor's first strike grant.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "94b86278-bfa6-4be4-8114-09b6c6c85724",
		Name:         "Claws of Valakut",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttachedPer(1, 0, mountainsTheAurasControllerControls),
			GrantToAttached("first strike"),
		},
	})
}
