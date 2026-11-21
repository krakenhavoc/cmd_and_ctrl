package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hyena Umbra — Enchantment — Aura {W} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has first strike.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     HyenaUmbraOracleID,
		Name:         "Hyena Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("first strike"),
		},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const HyenaUmbraOracleID = "354affb3-1be7-4a9e-91da-2989eb54f62d"
