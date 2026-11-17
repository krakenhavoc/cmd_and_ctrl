package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sandstorm — Instant {G} (Arabian Nights, #5641):
//
//	"Sandstorm deals 1 damage to each attacking creature."
//
// Mass damage over the creatures attacking as it resolves, whoever
// controls them, dealt at once by the spell (damageEachMatching). It is
// a partial sweep: it spares every creature that is not attacking.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7bb41690-f8ec-462a-ba29-be453eb86fca",
		Name:         "Sandstorm",
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
