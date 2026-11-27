package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elvish Herder — Creature — Elf {G} 1/1 (Urza's Saga, #5624):
//
//	"{G}: Target creature gains trample until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7832b90a-23b9-448c-9019-8ee2e8dfab31",
		Name:         "Elvish Herder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Target creature gains trample until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")),
			Cost:    effects.ManaCost("{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB726TargetGainsKeywordUntilEOT("Elvish Herder", "trample"),
		}},
	})
}
