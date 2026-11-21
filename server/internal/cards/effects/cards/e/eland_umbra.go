package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Eland Umbra — Enchantment — Aura {1}{W} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +0/+4.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     ElandUmbraOracleID,
		Name:         "Eland Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(0, 4)},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const ElandUmbraOracleID = "836fdd4b-98f8-4615-ae81-9bf63205c37d"
