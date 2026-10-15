package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eel Umbra — Enchantment — Aura {1}{U} (CR 702.89, #3654):
//
//	"Flash
//	 Enchant creature
//	 Enchanted creature gets +1/+1.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// Flash is a printed keyword on the Spec, as Airtight Alibi's is. No
// simplification.
func init() {
	Register(Spec{
		OracleID:        eelUmbraOracleID,
		Name:            "Eel Umbra",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Targets:         EnchantCreature(),
		Purpose:         game.Purpose{Aura: game.AuraBenefit},
		Static:          []game.StaticAbility{PumpAttached(1, 1)},
		Replacements:    []game.ReplacementEffect{UmbraArmor()},
	})
}

const eelUmbraOracleID = "4be1ecd0-ee41-4ac0-9155-0fcae7a4d337"
