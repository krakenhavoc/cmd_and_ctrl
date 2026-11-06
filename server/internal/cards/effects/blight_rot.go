package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blight Rot — Instant {2}{B} (Lorwyn Eclipsed, #4458):
//
//	"Put four -1/-1 counters on target creature."
//
// Plague Belcher's counter body at four, on any creature: the counters go
// on through AddCounter, so a counter doubler or a Vizier of Remedies
// sees them as it would any other -1/-1 counter. A target that has left
// in response gets nothing (CR 608.2b). The counters stay after the turn.
// This is not blight (the creature is targeted, and the caster need not
// control it).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cf5b1cf9-dcd3-44d1-b8b2-338651ebb2a1",
		Name:         "Blight Rot",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return AddCounter{Target: id, Kind: game.CounterMinusOne, N: 4}.Apply(ctx)
		},
	})
}
