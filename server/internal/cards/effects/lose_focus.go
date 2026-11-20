package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lose Focus — Instant {1}{U} (Modern Horizons 2, #5501):
//
//	"Replicate {U} (When you cast this spell, copy it for each time you
//	 paid its replicate cost. You may choose new targets for the copies.)
//	 Counter target spell unless its controller pays {2}."
//
// Spell Pierce's pay-unless on any spell, with Consign to Memory's
// replicate (replicate.go, CR 702.56a): each copy may choose a new
// target (CR 707.10c) and asks for its own {2}.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "1cea6439-7ae5-4887-8c33-7da9fb36e2d4",
		Name:          "Lose Focus",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{replicateMana("{U}", 20)},
		Triggered:     []game.TriggeredAbility{Replicate()},
		Targets:       TargetSpell("target spell"),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return CounterUnlessPaid{
				StackID:  item.Targets[0].ID,
				Cost:     "{2}",
				Question: "Lose Focus — pay {2} or your spell is countered",
			}.Apply(ctx)
		},
	})
}
