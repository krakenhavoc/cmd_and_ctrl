package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hyena Umbra — Enchantment — Aura {W} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has first strike.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     hyenaUmbraOracleID,
		Name:         "Hyena Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			PumpAttached(1, 1),
			GrantToAttached("first strike"),
		},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const hyenaUmbraOracleID = "354affb3-1be7-4a9e-91da-2989eb54f62d"
