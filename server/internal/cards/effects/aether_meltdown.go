package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Aether Meltdown — Enchantment — Aura {1}{U}:
//
//	"Flash (You may cast this spell any time you could cast an instant.)
//	 Enchant creature or Vehicle
//	 When this Aura enters, you get {E}{E} (two energy counters).
//	 Enchanted creature gets -4/-0."
//
// ADR 0129 PR 1. The enchant clause admits a Vehicle that is not a
// creature, and the CR 704.5m re-check runs the same clause, so the Aura
// stays on an uncrewed Vehicle. "Enchanted creature gets -4/-0" applies
// only while the host is a creature (pumpEnchantedCreature).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "433a2880-d989-4c40-99fa-0896de341b88",
		Name:            "Aether Meltdown",
		Completeness:    CompletenessFull,
		Purpose:         game.Purpose{Energy: 2},
		PrintedKeywords: []string{"flash"},
		Targets:         TargetPermanent("enchant creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
		Static:          []game.StaticAbility{pumpEnchantedCreature(-4, 0)},
		Triggered:       []game.TriggeredAbility{WhenThisEntersYouGetEnergy("Aether Meltdown", 2)},
	})
}
