package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     LionUmbraOracleID,
		Name:         "Lion Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(effects.ModifiedCreature),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("reach", "vigilance"),
		},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const LionUmbraOracleID = "95e8e6bb-d83b-46fd-bfdf-fc042d8edd4a"
