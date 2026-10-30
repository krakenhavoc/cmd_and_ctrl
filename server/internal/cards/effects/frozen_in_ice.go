package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frozen in Ice — Enchantment — Aura {2}{U}:
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature loses all abilities and can't become untapped."
//
// Infinite Coursework's shape with the stronger lock (#4574): "can't
// become untapped" holds against every untap, not only the controller's
// untap step, so an untap effect does nothing to the creature either (CR
// 614.17). It is an UntapStepRestriction with CantBecomeUntapped, read
// live from the Aura, so it ends the moment the Aura leaves. The entry
// trigger reads the Aura's attachment as it resolves
// (tapEnchantedCreatureOnEntry); the tap itself does not end with the
// Aura.
func init() {
	Register(Spec{
		OracleID:              "82fa14d8-5019-4e1b-a86c-923d9292b87a",
		Name:                  "Frozen in Ice",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedCantBecomeUntapped()},
		Static:                []game.StaticAbility{LoseAllAbilities()},
		Triggered:             []game.TriggeredAbility{WhenThisEnters("Frozen in Ice — tap enchanted creature", tapEnchantedCreatureOnEntry)},
	})
}
