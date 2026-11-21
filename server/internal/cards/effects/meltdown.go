package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Meltdown — Sorcery {X}{R} (Modern Horizons 3):
//
//	"Destroy each artifact with mana value X or less."
//
// Dauntless Dismantler's sweep without the "or more" half: one simultaneous
// destroy over every artifact whose mana value is at most the X chosen as the
// spell was cast. On the battlefield an {X} in a cost counts as 0 (CR 202.3e)
// and a token's mana value is 0, so X = 0 destroys tokens and free artifacts.
// It hits your own artifacts too, and an indestructible one survives.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8fd85a54-f0c7-4e9f-a4b4-b9b513b98d95",
		Name:         "Meltdown",
		XMatters:     true,
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: And(Artifact(), ManaValueLE(ctx.X()))}.Apply(ctx)
		},
	})
}
