package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Plummet — Instant {1}{G} (Archenemy, #5580):
//
//	"Destroy target creature with flying."
//
// Crushing Canopy's first mode as the whole spell: the pick is a creature
// with flying, and a creature that has lost flying in response is an
// illegal target, so the spell does nothing (CR 608.2b). The destroy is the
// plain DestroyTarget, so regeneration and indestructible apply.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "85bde6ac-3dd4-4946-8b57-24f57e3eae2b",
		Name:         "Plummet",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature with flying", effects.HasKeyword("flying")),
		OnResolve:    effects.DestroyTheTargetPermanent,
	})
}
