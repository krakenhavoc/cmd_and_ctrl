package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sulfurous Blast — Instant {2}{R}{R} (Commander Masters, #5948):
//
//	"Sulfurous Blast deals 2 damage to each creature and each player. If you
//	 cast this spell during your main phase, Sulfurous Blast deals 3 damage
//	 to each creature and each player instead."
//
// A spell cast in a main phase resolves in that same phase (the phase cannot
// end with the spell on the stack), so the phase is read as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f2b4f37b-270b-4746-83b3-51ff88ee3491",
		Name:         "Sulfurous Blast",
		Completeness: CompletenessFull,
		// ADR 0126 §6: two damage to every creature is the floor.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := 2
			if b14InYourMainPhase(ctx.Game, ctx.Controller()) {
				n = 3
			}
			return b23DamageEachCreatureAndEachPlayer(ctx, n)
		},
	})
}
