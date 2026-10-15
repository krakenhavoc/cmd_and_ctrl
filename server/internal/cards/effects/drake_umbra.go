package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drake Umbra — Enchantment — Aura {4}{U} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has flying.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     drakeUmbraOracleID,
		Name:         "Drake Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(3, 3),
			GrantToAttached("flying"),
		},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const drakeUmbraOracleID = "2b143ef1-afbb-499c-9189-1009105b406b"
