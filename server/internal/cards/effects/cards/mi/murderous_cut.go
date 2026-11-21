package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Murderous Cut — Instant {4}{B}:
//
//	"Delve. Destroy target creature."
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0ce7d5a-68fd-41f3-b5ba-a1a178cee9ec",
		Name:         "Murderous Cut",
		Completeness: effects.CompletenessFull,
		Delve:        true,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
