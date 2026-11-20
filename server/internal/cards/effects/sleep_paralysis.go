package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sleep Paralysis — Enchantment — Aura {3}{U}
// (Shadows over Innistrad, #5543):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Claustrophobia's text exactly: the enters trigger taps the enchanted
// creature (tapEnchantedCreatureOnEntry) and an untap-step restriction
// holds it (enchantedDoesntUntap).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "5a8bd4a4-8b10-4eef-8341-28be2afa2c00",
		Name:                  "Sleep Paralysis",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{WhenThisEnters("Sleep Paralysis — tap enchanted creature", tapEnchantedCreatureOnEntry)},
	})
}
