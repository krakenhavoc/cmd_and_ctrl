package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crab Umbra — Enchantment — Aura {U} (CR 702.89, #3654):
//
//	"Enchant creature
//	 {2}{U}: Untap enchanted creature.
//	 Umbra armor (If enchanted creature would be destroyed, instead remove
//	 all damage from it and destroy this Aura.)"
//
// The activated ability is the Aura's, not the creature's, and targets
// nothing: it finds the creature through the attachment, as Freed from
// the Real's untap does. No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     CrabUmbraOracleID,
		Name:         "Crab Umbra",
		Completeness: effects.CompletenessFull,
		Targets:      effects.EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{U}: Untap enchanted creature",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    effects.ManaCost("{2}{U}"),
			Effect:  effects.UntapAttachedHost,
		}},
		Replacements: []game.ReplacementEffect{effects.UmbraArmor()},
	})
}

const CrabUmbraOracleID = "7d3d1c90-a5fb-44ff-a85b-27c037ca6753"
