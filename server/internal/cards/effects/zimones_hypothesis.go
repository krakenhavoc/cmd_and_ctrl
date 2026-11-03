package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Zimone's Hypothesis — Instant {3}{U}{U} (Secrets of Strixhaven
// Commander, #4834):
//
//	"You may put a +1/+1 counter on a creature. Then choose odd or even.
//	 Return each creature with power of the chosen quality to its owner's
//	 hand. (Zero is even.)"
//
// Nothing is targeted. The controller picks up to one creature on the
// battlefield, anyone's, to take the counter (a PermanentsPick, so a
// table waiting on the answer restores), and the counter is on it before
// the next sentence starts: the parity question is asked after the counter
// has landed, and the power it is judged by is the power with that counter
// (CR 608.2c). Then the controller chooses odd or even, and every
// creature on the battlefield whose power has that parity goes back to its
// owner's hand as one return, the controller's own included. Power is read
// as the creatures are returned, without being clamped at zero (CR
// 107.1b), so a creature with -1 power is odd and one with power 0 is
// even, as the card says. With no creature on the battlefield nothing is asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ae27fa09-ab73-44b4-8e86-44d626ec221d",
		Name:         "Zimone's Hypothesis",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepBounce, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ChoosePermanents{
				Question: "Zimone's Hypothesis — you may put a +1/+1 counter on a creature",
				Candidates: func(g *game.Game, _ uuid.UUID) ([]uuid.UUID, int, int) {
					var ids []uuid.UUID
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.IsCreature() {
							ids = append(ids, c.InstanceID)
						}
					}
					if len(ids) == 0 {
						return nil, 0, 0
					}
					return ids, 0, 1
				},
				Keyed: zimonesHypothesisPick,
			}.Apply(ctx)
		},
	})
}

// zimonesHypothesisRecord files the creature that takes the counter.
const zimonesHypothesisRecord = "zimones-hypothesis/counter"

// zimonesHypothesisPick is the "you may": the rest of the card runs as
// steps, so the parity question waits for the counter. An on-disk key.
var zimonesHypothesisPick = RegisterPermanentsPick("prompt-run/zimones-hypothesis-counter", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		ctx.RecordStep(zimonesHypothesisRecord, picked.Cards())
		return zimonesHypothesisRest.Run(ctx)
	},
})

// zimonesHypothesisRest is the counter, then "choose odd or even" and the
// return.
var zimonesHypothesisRest = RegisterThenSteps("steps/zimones-hypothesis-rest",
	Instruction(func(ctx *Context) error {
		for _, id := range ctx.StepRecord(zimonesHypothesisRecord).IDs {
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
				return err
			}
		}
		return nil
	}),
	Instruction(func(ctx *Context) error {
		return PickOption{
			Question: "Zimone's Hypothesis — choose odd or even",
			Options:  []game.ChoiceOption{{Label: "Odd"}, {Label: "Even"}},
			ThenKey:  zimonesHypothesisParity,
		}.Apply(ctx)
	}),
)

// zimonesHypothesisParity returns each creature whose power has the chosen
// parity. An on-disk key.
var zimonesHypothesisParity = OptionPickThen("option-pick/zimones-hypothesis-parity", func(ctx *Context, r game.OptionPicked) error {
	if r.Index < 0 {
		return nil
	}
	odd := r.Index == 0
	return BounceAllMatching{
		Match: func(_ *game.Game, _ uuid.UUID, c game.Card) bool {
			return c.IsCreature() && (c.PowerForComparison()%2 != 0) == odd
		},
	}.Apply(ctx)
})
