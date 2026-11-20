package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Whipflare — Sorcery {1}{R} (New Phyrexia, #5575):
//
//	"Whipflare deals 2 damage to each nonartifact creature."
//
// Mass damage over the nonartifact creatures as it resolves, whoever
// controls them, dealt at once by the spell (damageEachMatching). It is a
// partial sweep: artifact creatures are spared.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a6b599db-8d45-4a11-9765-ae7eb235b4b2",
		Name:         "Whipflare",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 2, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, And(Creature(), Not(Artifact())), 2)
		},
	})
}
