package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     IndrikUmbraOracleID,
		Name:         "Indrik Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(4, 4),
			effects.GrantToAttached("first strike"),
			effects.Labelled(effects.BlockRequirementWhere(game.BlockRequirementLure, effects.AttachedToSource),
				"All creatures able to block enchanted creature do so."),
		},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const IndrikUmbraOracleID = "b585c218-1793-49a0-8a16-cb23b6efcc3c"
