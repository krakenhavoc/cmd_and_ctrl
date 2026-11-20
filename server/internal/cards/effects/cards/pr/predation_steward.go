package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Predation Steward — Creature — Phyrexian Elf Warrior {1}{G}, 2/2
// (Phyrexia: All Will Be One):
//
//	"This creature enters with two oil counters on it.
//	 {2}{G}, {T}, Remove an oil counter from this creature: Target creature
//	 gets +2/+2 until end of turn. Activate only as a sorcery."
//
// Any creature, yours or not, itself included.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e3d2b03-4c44-49a3-ab11-871ca9c372f7",
		Name:         "Predation Steward",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.B10EntersWithCounters(effects.OilCounter, 2, "Predation Steward: enters with two oil counters on it")},
		Activated: []effects.ActivatedAbility{{
			Label:        "{2}{G}, {T}, Remove an oil counter from this creature: Target creature gets +2/+2 until end of turn. Activate only as a sorcery.",
			Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
			Cost:         effects.Plus(effects.ManaCost("{2}{G}"), effects.TapCost(), effects.RemoveCountersFromThis(effects.OilCounter, 1)),
			Targets:      effects.TargetCreature("target creature"),
			SorcerySpeed: true,
			Effect:       effects.BoostTargetUntilEOT("Predation Steward — +2/+2", 2, 2),
		}},
	})
}
