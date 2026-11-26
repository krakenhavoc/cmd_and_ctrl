package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Whirlwind — Sorcery {2}{G}{G} (Commander 2014):
//
//	"Destroy all creatures with flying."
//
// Flying is read through the layered view as the spell resolves, so a
// creature that has gained or lost flying is judged by what it has now.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3946df7-cb24-47c0-a8b0-1bdf95d99926",
		Name:         "Whirlwind",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.HasKeyword("flying"))}.Apply(ctx)
		},
	})
}
