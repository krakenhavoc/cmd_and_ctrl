package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drake Umbra — Enchantment — Aura {4}{U} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +3/+3 and has flying.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     DrakeUmbraOracleID,
		Name:         "Drake Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(3, 3),
			effects.GrantToAttached("flying"),
		},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const DrakeUmbraOracleID = "2b143ef1-afbb-499c-9189-1009105b406b"
