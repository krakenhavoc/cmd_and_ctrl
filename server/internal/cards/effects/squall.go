package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Squall — Sorcery {2}{G} (Starter 1999, #5620):
//
//	"Squall deals 2 damage to each creature with flying."
//
// Firespout's green half on its own: one batch of damage from the spell
// (damageEachMatching), to each creature that has flying through the
// layers as it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "16fe0abe-04fa-4023-8ee0-043a076e78bb",
		Name:         "Squall",
		Completeness: CompletenessFull,
		// ADR 0126 §6: only creatures with flying.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), HasKeyword("flying")), 2)
		},
	})
}
