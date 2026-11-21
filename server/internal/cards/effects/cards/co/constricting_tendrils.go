package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Constricting Tendrils — Instant {U} (Conflux, #5587):
//
//	"Target creature gets -3/-0 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Shrink at -3/-0 through the shared boostTheTargetUntilEOT body, ending
// in the cleanup step (CR 514.2), with the engine's cycling row (CR
// 702.29a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9e34ab18-5d39-49d9-9d3d-e1d4a3611b87",
		Name:         "Constricting Tendrils",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.HarmUntilEOT(0, -3, 0)),
		OnResolve:    effects.BoostTheTargetUntilEOT(-3, 0, "Constricting Tendrils — -3/-0 until end of turn"),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
