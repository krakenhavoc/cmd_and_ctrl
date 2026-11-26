package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Primal Bellow — Instant {G} (Zendikar, S82 OLD b496):
//
//	"Target creature gets +1/+1 until end of turn for each Forest you control."
//
// The Forests are counted as the spell resolves (CR 608.2h), by land type, a word
// of the card's text (CR 612.2); a target that is no longer legal does nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5cd064b9-1277-4edc-bf2c-82eb5e62f9d7",
		Name:         "Primal Bellow",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.Old496PumpTargetPerForest,
	})
}
