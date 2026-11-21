package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boar Umbra — Enchantment — Aura {2}{G} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     BoarUmbraOracleID,
		Name:         "Boar Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static:       []game.StaticAbility{effects.PumpAttached(3, 3)},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const BoarUmbraOracleID = "75e4448b-5bd7-4ed4-8db8-820cd6d04c12"
