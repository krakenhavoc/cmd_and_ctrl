package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fyndhorn Bow — Artifact {2} (Ice Age, S82 OLD, #5636):
//
//	"{3}, {T}: Target creature gains first strike until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d37fb8bf-d293-4ac3-b744-74e4caade975",
		Name:         "Fyndhorn Bow",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Target creature gains first strike until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "first strike")),
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.S82OldB726TargetGainsKeywordUntilEOT("Fyndhorn Bow", "first strike"),
		}},
	})
}
