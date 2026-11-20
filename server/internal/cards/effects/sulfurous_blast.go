package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sulfurous Blast — Instant {2}{R}{R} (Time Spiral, #5596):
//
//	"Sulfurous Blast deals 2 damage to each creature and each player. If
//	 you cast this spell during your main phase, Sulfurous Blast deals 3
//	 damage to each creature and each player instead."
//
// Cave-In's body at 2, or at 3 when it was cast in its controller's main
// phase (tspB2MainPhaseAmount, read as it resolves).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2b4f37b-270b-4746-83b3-51ff88ee3491",
		Name:         "Sulfurous Blast",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b23DamageEachCreatureAndEachPlayer(ctx, tspB2MainPhaseAmount(ctx, 2, 3))
		},
	})
}
