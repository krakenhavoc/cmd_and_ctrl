package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marrow Shards — Instant {W/P} (New Phyrexia, #5575):
//
//	"({W/P} can be paid with either {W} or 2 life.)
//	 Marrow Shards deals 1 damage to each attacking creature."
//
// Sandstorm's mass damage over the creatures attacking as it resolves,
// whoever controls them (damageEachMatching). The Phyrexian symbol is
// paid by the engine at cast (CR 107.4f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be2ae8dc-afae-4bcf-8a97-08a2764d1148",
		Name:         "Marrow Shards",
		Completeness: CompletenessFull,
		Purpose: game.Purpose{
			Answers: game.AnswerRemove,
			Sweep:   game.Sweep{Matches: game.SweepCreatures, How: game.SweepDamage, Amount: 1, Partial: true},
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachMatching(ctx, AttackingCreature(), 1)
		},
	})
}
