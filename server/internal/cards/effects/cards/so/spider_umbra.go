package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spider Umbra — Enchantment — Aura {G} (CR 702.89, #3654):
//
//	"Enchant creature
//	 Enchanted creature gets +1/+1 and has reach.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     SpiderUmbraOracleID,
		Name:         "Spider Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Static: []game.StaticAbility{
			effects.PumpAttached(1, 1),
			effects.GrantToAttached("reach"),
		},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const SpiderUmbraOracleID = "f32aaa81-96bd-4213-b0b8-8c021f3cea57"
