package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Apostle's Blessing — Instant {1}{W/P}:
//
//	"({W/P} can be paid with either {W} or 2 life.)
//	 Target artifact or creature you control gains protection from artifacts
//	 or from the color of your choice until end of turn."
//
// The choice (artifacts, or one of the five colours) is made as the spell
// resolves; a target that left in response is asked nothing.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ea393815-0202-4edc-aa1e-6885fe9b20ab",
		Name:         "Apostle's Blessing",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection")),
		Targets:      effects.TargetPermanent("target artifact or creature you control", effects.Or(effects.Artifact(), effects.Creature()), effects.YouControl()),
		OnResolve:    effects.ApostlesBlessingProtection,
	})
}
