package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Runner's Bane — Enchantment — Aura {1}{U} (Dragon's Maze, S82 OLD b405,
// #5564):
//
//	"Enchant creature with power 3 or less
//	 When this Aura enters, tap enchanted creature.
//	 Enchanted creature doesn't untap during its controller's untap step."
//
// Bubble Snare's entry tap and untap restriction. The enchant clause is the
// power gate, read at announce and by the CR 704.5m legality check, so an
// Aura on a creature whose power rises above 3 falls off.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:              "31c30c4f-12f4-4cbc-b7c1-123f36b8df75",
		Name:                  "Runner's Bane",
		Completeness:          effects.CompletenessFull,
		Targets:               effects.EnchantCreature(effects.PowerLE(3)),
		Purpose:               game.Purpose{Aura: game.AuraHarm},
		UntapStepRestrictions: []game.UntapStepRestriction{effects.EnchantedDoesntUntap()},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Runner's Bane — tap enchanted creature", effects.TapEnchantedCreatureOnEntry),
		},
	})
}
