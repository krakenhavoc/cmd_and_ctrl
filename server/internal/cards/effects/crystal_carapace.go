package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crystal Carapace — Enchantment — Aura {3}{G} (March of the Machine):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has ward {2}.
//	 Cycling {2}"
//
// The pump is PumpAttached; the ward is WardAttached, as on Chains of Custody:
// ward is a triggered ability with a cost (CR 702.21a), so the grant is the Aura
// carrying a ward trigger that watches the permanent it is attached to.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "06a6130f-f27b-4db0-915d-decd9a7d3c15",
		Name:         "Crystal Carapace",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		Triggered: []game.TriggeredAbility{
			WardAttached(WardMana("{2}"), "Crystal Carapace — ward {2}"),
		},
		Activated: []ActivatedAbility{Cycling("{2}")},
	})
}
