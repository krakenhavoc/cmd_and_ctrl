package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Filter Out — Instant {1}{U}{U} (March of the Machine: The Aftermath):
//
//	"Return all noncreature, nonland permanents to their owners'
//	 hands."
//
// Devastation Tide's BounceAllMatching with a narrower filter, read on the
// battlefield as the spell resolves: a permanent that is also a creature
// (an artifact creature, an animated Vehicle or Aura-turned-creature) is
// left alone, lands are left alone, and tokens that go to a hand cease to
// exist. Symmetric: your own permanents return too. The sweep is declared
// Partial (it spares creatures and lands, ADR 0126).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8c95d823-6bbd-4484-8f38-b1bb20403085",
		Name:         "Filter Out",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.And(effects.Nonland(), effects.Noncreature())}.Apply(ctx)
		},
	})
}
