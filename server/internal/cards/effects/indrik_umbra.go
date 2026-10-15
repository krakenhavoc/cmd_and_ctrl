package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Indrik Umbra — Enchantment — Aura {4}{G}{W} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +4/+4 and has first strike, and all creatures
//	 able to block it do so.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// The block requirement is Lure's (BlockRequirementWhere over the
// attachment). No simplification.
func init() {
	Register(Spec{
		OracleID:     indrikUmbraOracleID,
		Name:         "Indrik Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(4, 4),
			GrantToAttached("first strike"),
			Labelled(BlockRequirementWhere(game.BlockRequirementLure, AttachedToSource),
				"All creatures able to block enchanted creature do so."),
		},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const indrikUmbraOracleID = "b585c218-1793-49a0-8a16-cb23b6efcc3c"
