package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// By Force — Sorcery {X}{R} (Secret Lair Drop):
//
//	"Destroy X target artifacts."
//
// Heliod's Intervention's first mode on its own: the target count IS the announced X
// (CountFromX, pinned at announce and re-checked slot by slot as it resolves), and every
// target still legal is destroyed in one simultaneous event (DestroyEachLegalTarget), so an
// indestructible artifact survives and a target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	spec := effects.TargetPermanent("X target artifacts", effects.Artifact())
	spec.CountFromX = true
	effects.Register(effects.Spec{
		OracleID:     "7ce950fb-e065-45d3-aeff-f8c475460754",
		Name:         "By Force",
		Completeness: effects.CompletenessFull,
		Targets:      spec,
		OnResolve:    effects.DestroyEachLegalTarget,
	})
}
