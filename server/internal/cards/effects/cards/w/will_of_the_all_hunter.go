package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Will of the All-Hunter — Instant {1}{W} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target creature gets +2/+2 until end of turn. If it's blocking, instead
//	 put two +1/+1 counters on it.
//	 Cycling {2}"
//
// Whether the creature is blocking is read as the spell resolves (CR 506.4):
// a blocker gets two counters that stay, anything else the pump that ends at
// cleanup. A target that has gone does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dec281b4-e6d6-4b78-b755-358e07eb7c06",
		Name:         "Will of the All-Hunter",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
		OnResolve:    effects.ResolveSteps(effects.S82OldB114WillOfTheAllHunter()),
	})
}
