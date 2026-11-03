package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blossombind — {1}{U} Enchantment — Aura (Lorwyn Eclipsed, #4458):
//
//	"Enchant creature
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature can't become untapped and can't have counters
//	 put on it."
//
// Frozen in Ice without the ability loss: the entry tap and "can't become
// untapped" are its UntapStepRestriction with CantBecomeUntapped (#4574,
// CR 614.17). "Can't have counters put on it" is #4692's prohibition
// (CR 101.2, game/cant_have_counters.go): no counter is put on the
// creature by any effect, cost, damage result or proliferate, and no
// counter replacement sees the attempt. Both are read live from the Aura,
// so they end when it leaves; the counters already on the creature stay.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:              "0ecb35b3-5bd1-4467-a4a3-ce468023ced2",
		Name:                  "Blossombind",
		Completeness:          CompletenessFull,
		Targets:               EnchantCreature(),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{enchantedCantBecomeUntapped()},
		Static:                []game.StaticAbility{AttachedCantHaveCounters()},
		Triggered:             []game.TriggeredAbility{WhenThisEnters("Blossombind — tap enchanted creature", tapEnchantedCreatureOnEntry)},
	})
}
