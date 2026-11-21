package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Drag Down — Instant {2}{B} (Conflux, #5587):
//
//	"Domain — Target creature gets -1/-1 until end of turn for each
//	 basic land type among lands you control."
//
// The count is the shared DomainCount (domain.go, CR 305.6), read once
// as the spell resolves (CR 608.2h). With no basic land type it does
// nothing; a creature brought to 0 toughness dies to the state-based
// check (CR 704.5f).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30e397e9-a705-4484-8387-04fb84010b2d",
		Name:         "Drag Down",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, 0, 0)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ConB1DomainBoostTheTarget(-1, "Drag Down — -1/-1 for each basic land type"),
	})
}
