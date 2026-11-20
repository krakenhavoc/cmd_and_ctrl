package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "3ac01298-e45b-4ea0-8379-2074279716a4",
		Name:         "Citywide Bust",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(DestroyAllMatching{Match: And(Creature(), ToughnessGE(4))}),
	})
}
