package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ravnica at War — Sorcery {3}{W} (War of the Spark, #5520):
//
//	"Exile all multicolored permanents."
//
// A mass exile (ExileAllMatching) over the permanents with two or more colors
// as the spell resolves, everyone's, yours included. Exile is not destruction,
// so indestructible does not save them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0012bc78-e69d-4a67-a302-e5fe0dfd4407",
		Name:         "Ravnica at War",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepAllPermanents, How: game.SweepExile, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileAllMatching{Match: effects.Multicolored()}.Apply(ctx)
		},
	})
}
