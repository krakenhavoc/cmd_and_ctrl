package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Feral Ferocity — Instant {1}{G} (Marvel Super Heroes, #4455):
//
//	"Target creature gets +4/+4 until end of turn."
//
// Titanic Growth under a Marvel name: the target clause is plain "target
// creature", so an opponent's creature is a legal target, and the boost
// is plusFourPlusFourToTheTarget.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5e968bf3-8d83-480b-81a4-98883cafd04f",
		Name:         "Feral Ferocity",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.PlusFourPlusFourToTheTarget("Feral Ferocity"),
	})
}
