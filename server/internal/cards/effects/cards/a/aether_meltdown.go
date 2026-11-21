package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "433a2880-d989-4c40-99fa-0896de341b88",
		Name:            "Aether Meltdown",
		Completeness:    effects.CompletenessFull,
		Purpose:         game.Purpose{Energy: 2},
		PrintedKeywords: []string{"flash"},
		Targets:         effects.TargetPermanent("enchant creature or Vehicle", effects.Or(effects.Creature(), effects.HasSubtype("Vehicle"))),
		Static:          []game.StaticAbility{effects.PumpEnchantedCreature(-4, 0)},
		Triggered:       []game.TriggeredAbility{effects.WhenThisEntersYouGetEnergy("Aether Meltdown", 2)},
	})
}
