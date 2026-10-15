package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Boar Umbra — Enchantment — Aura {2}{G} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     boarUmbraOracleID,
		Name:         "Boar Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{PumpAttached(3, 3)},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const boarUmbraOracleID = "75e4448b-5bd7-4ed4-8db8-820cd6d04c12"
