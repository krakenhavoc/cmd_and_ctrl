package cardsth

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Their Name Is Death — Sorcery {3}{B}{B}{B} (Warhammer 40,000 Commander,
// #5948):
//
//	"Destroy all nonartifact creatures."
//
// Damnation's body over nonartifact creatures: an artifact creature is
// spared, a regeneration shield still saves what it covers (CR 701.19).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "97f9aa5e-e34d-4029-9213-b32ebaf9859c",
		Name:         "Their Name Is Death",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.Not(effects.Artifact()))}.Apply(ctx)
		},
	})
}
