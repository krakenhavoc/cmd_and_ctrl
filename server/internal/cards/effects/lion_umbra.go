package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lion Umbra — Enchantment — Aura {G}{G} (CR 702.89, #3654):
//
//	"Enchant modified creature (Equipment, Auras its controller controls,
//	 and counters are modifications.)
//	 Enchanted creature gets +3/+3 and has reach and vigilance.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// "Modified" is Lethal Throwdown's and Kodama's predicate (CR 700.9).
// No simplification.
func init() {
	Register(Spec{
		OracleID:     lionUmbraOracleID,
		Name:         "Lion Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(modifiedCreature),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			GrantToAttached("reach", "vigilance"),
		},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const lionUmbraOracleID = "95e8e6bb-d83b-46fd-bfdf-fc042d8edd4a"
