package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mammoth Umbra — Enchantment — Aura {4}{W} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has vigilance.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     mammothUmbraOracleID,
		Name:         "Mammoth Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			GrantToAttached("vigilance"),
		},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const mammothUmbraOracleID = "cd0e6cdb-9b7f-47d6-9478-fb50003165af"
