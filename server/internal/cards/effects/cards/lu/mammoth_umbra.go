package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mammoth Umbra — Enchantment — Aura {4}{W} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has vigilance.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     MammothUmbraOracleID,
		Name:         "Mammoth Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("vigilance"),
		},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const MammothUmbraOracleID = "cd0e6cdb-9b7f-47d6-9478-fb50003165af"
