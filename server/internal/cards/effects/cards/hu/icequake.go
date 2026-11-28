package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Icequake — Sorcery {1}{B}{B} (Ice Age, S82 OLD, #5636):
//
//	"Destroy target land. If that land was a snow land, Icequake deals 1 damage to that land's controller."
//
// The land is read before it is destroyed (snow, and who controlled it); the damage does not depend on
// the destruction succeeding.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c0a6ac1-b1cc-405a-95fb-f06b7fecbb54",
		Name:         "Icequake",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPermanent("target land", effects.Land()),
		OnResolve:    effects.S82OldB842Icequake,
	})
}
