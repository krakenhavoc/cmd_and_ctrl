package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     crabUmbraOracleID,
		Name:         "Crab Umbra",
		Completeness: CompletenessFull,
		Targets:      EnchantCreature(),
		Purpose:      game.Purpose{Aura: game.AuraBenefit},
		Activated: []ActivatedAbility{{
			Label:   "{2}{U}: Untap enchanted creature",
			Purpose: game.Purpose{Answers: game.AnswerCombatGrant},
			Cost:    ManaCost("{2}{U}"),
			Effect:  untapAttachedHost,
		}},
		Replacements: []game.ReplacementEffect{UmbraArmor()},
	})
}

const crabUmbraOracleID = "7d3d1c90-a5fb-44ff-a85b-27c037ca6753"
