package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stave Off — Instant {W} (Magic 2012, #5573):
//
//	"Target creature gains protection from the color of your choice
//	 until end of turn."
//
// Blessed Breath's body (protectionFromAChosenColorForTheTarget) on any
// target creature: the color is chosen as it resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c77c23bc-efde-43ce-bf6f-7a8531d81d89",
		Name:         "Stave Off",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection")),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ProtectionFromAChosenColorForTheTarget("Stave Off"),
	})
}
