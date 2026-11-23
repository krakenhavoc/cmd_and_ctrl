package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Furnace Strider — Creature — Phyrexian Beast {4}{R}, 4/5 (Phyrexia: All
// Will Be One):
//
//	"This creature enters with two oil counters on it.
//	 Remove an oil counter from this creature: Target creature you control
//	 gains haste until end of turn."
//
// The target is a creature its controller controls, itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85488433-3c23-4a29-b063-53dba71f26c4",
		Name:         "Furnace Strider",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 2, "Furnace Strider: enters with two oil counters on it")},
		Activated: []effects.ActivatedAbility{{
			Label:   "Remove an oil counter from this creature: Target creature you control gains haste until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "haste")),
			Cost:    effects.RemoveCountersFromThis(effects.OilCounter, 1),
			Targets: effects.TargetCreature("target creature you control", effects.YouControl()),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("haste", "Furnace Strider — haste until end of turn"),
		}},
	})
}
