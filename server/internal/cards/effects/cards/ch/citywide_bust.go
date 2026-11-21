package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Citywide Bust — Sorcery {1}{W}{W} (Guilds of Ravnica, #5525):
//
//	"Destroy all creatures with toughness 4 or greater."
//
// Retribution of the Meek's sweep read on toughness: toughness is read
// post-layer as it resolves, so a creature pumped to 4 is destroyed and
// one shrunk below it is spared.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3ac01298-e45b-4ea0-8379-2074279716a4",
		Name:         "Citywide Bust",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.ResolveSteps(effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.ToughnessGE(4))}),
	})
}
