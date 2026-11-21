package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessed Breath — Instant — Arcane {W}:
//
//	"Target creature you control gains protection from the color of your
//	choice until end of turn.
//	Splice onto Arcane {W}"
//
// Center Soul's body: the colour is chosen as the text resolves. Splice
// is ADR 0202 (#1989).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e83e1f99-1f9b-49fc-bdc7-5f6335e401e4",
		Name:         "Blessed Breath",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "protection")),
		Completeness: effects.CompletenessFull,
		Splice:       effects.SpliceOntoArcane("{W}"),
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		OnResolve:    effects.ProtectionFromAChosenColorForTheTarget("Blessed Breath"),
	})
}
