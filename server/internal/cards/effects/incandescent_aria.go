package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Incandescent Aria — Sorcery {R}{G}{W} (Streets of New Capenna, #5491):
//
//	"Incandescent Aria deals 3 damage to each nontoken creature."
//
// A board wipe that spares tokens, so its Sweep is partial.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d031b482-b33d-4eb4-a5bc-ea96e66a1d99",
		Name:         "Incandescent Aria",
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 3, Partial: true}},
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), Not(IsTokenPredicate())), 3)
		},
	})
}
