package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taste for Mayhem — Enchantment — Aura {R} (Dissension, #5598):
//
//	"Enchant creature
//	 Enchanted creature gets +2/+0.
//	 Hellbent — Enchanted creature gets an additional +2/+0 as long as you
//	 have no cards in hand."
//
// Fire-Rim Form's PumpAttached, and a second layer 7c pump that reads the
// Aura controller's hand (disAttachedPumpWhileHellbent).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60e7517f-7cc5-4073-9f2c-3e6acb891128",
		Name:         "Taste for Mayhem",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Targets:      EnchantCreature(),
		Static: []game.StaticAbility{
			PumpAttached(2, 0),
			disAttachedPumpWhileHellbent(2, 0),
		},
	})
}
