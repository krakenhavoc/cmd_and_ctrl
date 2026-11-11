package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Incremental Blight — Sorcery {3}{B}{B} (Lorwyn Eclipsed Commander,
// #5368):
//
//	"Put a -1/-1 counter on target creature, two -1/-1 counters on
//	 another target creature, and three -1/-1 counters on a third target
//	 creature."
//
// Three clauses, the second and third Distinct ("another", "a third",
// CR 601.2c). Each pick is checked separately at resolution (CR 608.2b):
// a pick that left gets nothing and the others still land.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "159cd205-6afc-4d8c-a4b9-501911a9c345",
		Name:         "Incremental Blight",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets: Clauses(
			TargetCreature("target creature"),
			Distinct(TargetCreature("another target creature")),
			Distinct(TargetCreature("a third target creature")),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for slot, n := range []int{1, 2, 3} {
				t, ok := ctx.ClauseTarget(slot)
				if !ok {
					continue
				}
				if err := (AddCounter{Target: t.ID, Kind: game.CounterMinusOne, N: n}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
