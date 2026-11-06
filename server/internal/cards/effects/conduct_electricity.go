package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Conduct Electricity — Instant {4}{R} (Bloomburrow, #4467):
//
//	"Conduct Electricity deals 6 damage to target creature and 2 damage
//	 to up to one target creature token."
//
// Arc Trail's shape (damageEachClauseTarget): two clauses, dealt as one
// damage instance. The second clause is optional, and it is a separate
// "target" word, so the same creature token may be chosen for both and
// is dealt 8 (CR 115.3). A clause whose target has become illegal is
// skipped (CR 608.2b); when both are illegal the spell does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0369627e-3b83-4e5b-a8af-a99f944443b3",
		Name:         "Conduct Electricity",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCreature("target creature"),
			TargetCreature("up to one target creature token", IsTokenPredicate()).WithCount(0, 1),
		),
		Purpose: game.Purpose{Targets: game.ForTargets(DamageToTarget(0, 6), DamageToTarget(1, 2)), Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return damageEachClauseTarget(ctx, 6, 2)
		},
	})
}
