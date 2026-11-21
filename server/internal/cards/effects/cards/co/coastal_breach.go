package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Coastal Breach — Sorcery {6}{U} (Commander 2016, #5537):
//
//	"Undaunted (This spell costs {1} less to cast for each opponent.)
//	 Return all nonland permanents to their owners' hands."
//
// Devastation Tide's bounce without the miracle, with Curtains' Call's
// undaunted (CR 702.125a): a self cost modifier counting the caster's
// opponents still in the game. Tokens are returned too and cease to
// exist (CR 111.8).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e42b6bd7-8fdf-4452-85cc-9aed790ff682",
		Name:         "Coastal Breach",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepNonlandPermanents, How: game.SweepBounce}},
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.OpponentsOfCaster(), "Undaunted"),
		},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.BounceAllMatching{Match: effects.Nonland()}.Apply(ctx)
		},
	})
}
