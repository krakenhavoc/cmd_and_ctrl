package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Watery Grasp — Enchantment — Aura {U}:
//
//	"Enchant creature
//	 Enchanted creature doesn't untap during its controller's untap step.
//	 Waterbend {5}: Enchanted creature's owner shuffles it into their
//	 library. (While paying a waterbend cost, you can tap your artifacts
//	 and creatures to help. Each one pays for {1}.)"
//
// Claustrophobia's untap lock, without its entry tap. The waterbend
// ability is Invasion Submersible's cost (`WaterbendCost`, CR 701.67a):
// {5}, and each artifact or creature tapped toward it pays {1}, even
// one that arrived this turn (the ruling). The enchanted creature is
// read as the ability resolves, or as the Aura last enchanted it if the
// Aura has gone (CR 608.2h); a creature that is no longer that object
// on the battlefield is left alone (CR 400.7). Its OWNER's library takes
// it and that owner shuffles, whoever controls the Aura or the creature.
//
// No simplifications. Avatar: The Last Airbender (#4459).
func init() {
	Register(Spec{
		OracleID:              "520b2799-66bd-40bf-9bc0-9cd3aed321f5",
		Name:                  "Watery Grasp",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedDoesntUntap()},
		Activated: []ActivatedAbility{{
			Label:   "Waterbend {5}: Enchanted creature's owner shuffles it into their library.",
			Purpose: game.Purpose{Answers: game.AnswerRemove},
			Cost:    WaterbendCost("{5}"),
			Effect:  avB7ShuffleEnchantedIntoLibrary,
		}},
	})
}
