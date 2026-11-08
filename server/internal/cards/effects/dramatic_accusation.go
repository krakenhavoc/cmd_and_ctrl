package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dramatic Accusation — Enchantment — Aura {2}{U} (Murders at Karlov
// Manor, #4469):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 {U}{U}: Shuffle enchanted creature into its owner's library."
//
// Claustrophobia's tap and untap lock, plus Watery Grasp's shuffle at
// {U}{U} and at instant speed. The creature goes to its OWNER's library
// and that owner shuffles, whoever controls the Aura or the creature. If
// the Aura has left by the time the ability resolves, the creature it
// last enchanted is the one shuffled (CR 608.2h); a creature that is no
// longer that object on the battlefield is left alone (CR 400.7). Only
// the Aura's controller may activate it (CR 602.2, the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "677b9c19-4886-4e06-88f8-9afa94e58aa8",
		Name:                  "Dramatic Accusation",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Dramatic Accusation — tap enchanted creature", tapEnchantedCreatureOnEntry),
		},
		Activated: []ActivatedAbility{{
			Label:   "{U}{U}: Shuffle enchanted creature into its owner's library.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    ManaCost("{U}{U}"),
			Effect:  avB7ShuffleEnchantedIntoLibrary,
		}},
	})
}
