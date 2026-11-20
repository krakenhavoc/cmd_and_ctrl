package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Needle Storm — Sorcery {2}{G} (Portal, #5630):
//
//	"Needle Storm deals 4 damage to each creature with flying."
//
// Tremor's sweep, turned round: only creatures with flying are hit.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f513d1e-03ca-4e0f-af48-6490d71d5c41",
		Name:         "Needle Storm",
		Completeness: CompletenessFull,
		// ADR 0126 §6: a creature without flying is spared.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 4, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), HasKeyword("flying")), 4)
		},
	})
}
