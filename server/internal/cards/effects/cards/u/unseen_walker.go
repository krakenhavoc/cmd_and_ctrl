package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unseen Walker — Creature — Dryad {1}{G}, 1/1 (Mirage, S82 OLD b808):
//
//	"Forestwalk (This creature can't be blocked as long as defending player
//	 controls a Forest.)
//	 {1}{G}{G}: Target creature gains forestwalk until end of turn."
//
// Forestwalk is the engine's (CR 702.14); the activation grants it to any
// creature, this one included, locked to the target as the ability resolves
// (CR 611.2c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f1768dd5-c5c8-4059-9695-55b8a11eeffd",
		Name:            "Unseen Walker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"forestwalk"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{G}{G}: Target creature gains forestwalk until end of turn.",
			Purpose: effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "forestwalk")),
			Cost:    effects.ManaCost("{1}{G}{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.OldB808TargetGainsForestwalk,
		}},
	})
}
