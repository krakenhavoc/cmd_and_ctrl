package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Towering Viewpoint — Artifact Creature — Wall {2}, 0/4 (Assassin's
// Creed, #5466):
//
//	"Defender, reach
//	 Leap of Faith — {3}: Target creature gains flying until end of
//	 turn."
//
// "Leap of Faith" is an ability word with no rules meaning.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "46c404c3-636a-4945-a61c-62f013a9c50f",
		Name:            "Towering Viewpoint",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender", "reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Leap of Faith — {3}: Target creature gains flying until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying")),
			Cost:    effects.ManaCost("{3}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.EbTargetCreatureGainsUntilEOT("flying", "Towering Viewpoint — flying until end of turn"),
		}},
	})
}
