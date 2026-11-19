package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Stony Strength — Instant {G} (Foundations Jumpstart):
//
//	"Put a +1/+1 counter on target creature you control. Untap that
//	 creature."
//
// The counter, then the untap, on the one target (CR 608.2c). A target
// that has gone by resolution fizzles the spell (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "da5a7cb8-1fa7-4b91-abd4-9e766291be63",
		Name:         "Stony Strength",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerPump},
		Targets:      TargetCreature("target creature you control", YouControl()),
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				id, ok := b16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				ctx.RecordStep(stonyStrengthRecord, []uuid.UUID{id})
				return AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
			}),
			Instruction(func(ctx *Context) error {
				if ids := ctx.StepRecord(stonyStrengthRecord).IDs; len(ids) > 0 {
					return UntapTarget{Target: ids[0]}.Apply(ctx)
				}
				return nil
			}),
		),
	})
}

// stonyStrengthRecord is the creature that got the counter.
const stonyStrengthRecord = "stony-strength/target"
