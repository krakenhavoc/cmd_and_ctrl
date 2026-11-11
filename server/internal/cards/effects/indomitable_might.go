package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Indomitable Might — Enchantment — Aura {3}{G} (Wilds of Eldraine
// Commander, #5368):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +3/+3.
//	 Enchanted creature's controller may have it assign its combat damage
//	 as though it weren't blocked."
//
// +3/+3 is a layer 7c pump. "May assign as though it weren't blocked" is
// Characteristic.AssignsAsThoughUnblocked (CR 510.1c, Wolverine, Claws
// Out), written onto the enchanted creature by a layer 6 static: the
// combat damage step asks the creature's controller (not the Aura's)
// whether to divide its damage among its blockers or assign all of it to
// what it attacks. It goes when the Aura does.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "069c3c87-81a8-4d2a-ad2f-8d3243d0e3db",
		Name:            "Indomitable Might",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			cmdr2B19AttachedAssignsAsThoughUnblocked(),
		},
	})
}
