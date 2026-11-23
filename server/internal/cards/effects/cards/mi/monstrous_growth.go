package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monstrous Growth — Sorcery {1}{G} (Portal, #5630):
//
//	"Target creature gets +4/+4 until end of turn."
//
// Titanic Growth's body (plusFourPlusFourToTheTarget), as a sorcery.
// The pump ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "35a05836-38d7-45c7-ac9a-996a682c2129",
		Name:         "Monstrous Growth",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.PlusFourPlusFourToTheTarget("Monstrous Growth"),
	})
}
