package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hour of Reckoning — Sorcery {4}{W}{W}{W} (EDHREC rank 1032):
//
//	"Convoke
//	 Destroy all nontoken creatures."
//
// The token deck's one-sided wrath: the tokens help pay for it and
// then survive it. Convoke is the S22 tap cost (Chord of Calling's),
// and the sweep is DestroyAllMatching over "creature that isn't a
// token" — one simultaneous event, so a Blood Artist sees every death
// at once.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "3bc13640-03f8-4b19-b0a4-7e7cb5271c0c",
		Name:     "Hour of Reckoning",
		// ADR 0126 §6: a token stays.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.Not(effects.IsTokenPredicate()))}.Apply(ctx)
		},
	})
}
