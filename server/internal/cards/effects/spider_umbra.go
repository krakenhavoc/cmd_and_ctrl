package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spider Umbra — Enchantment — Aura {G} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has reach.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     spiderUmbraOracleID,
		Name:         "Spider Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("reach"),
		},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const spiderUmbraOracleID = "f32aaa81-96bd-4213-b0b8-8c021f3cea57"
