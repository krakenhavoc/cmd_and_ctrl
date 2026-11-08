package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Starlight Snare — Enchantment — Aura {2}{U} (Foundations, #4465):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap
//	 step."
//
// Claustrophobia's text at a different cost: the enters trigger finds
// the creature through the Aura as it resolves, and the untap-step
// restriction holds only while the Aura is attached.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "f0691d46-5bf4-4617-8a0a-19d17afb7219",
		Name:                  "Starlight Snare",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered:             []game.TriggeredAbility{WhenThisEnters("Starlight Snare — tap enchanted creature", tapEnchantedCreatureOnEntry)},
	})
}
