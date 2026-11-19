package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Abandon Reason — {2}{R} Instant:
//
//	"Up to two target creatures each get +1/+0 and gain first strike until
//	end of turn.
//	Madness {1}{R}"
func init() {
	Register(Spec{
		OracleID:     "fa43793d-a215-4a64-a09c-ed9799f745a4",
		Name:         "Abandon Reason",
		Completeness: CompletenessFull,
		Madness:      "{1}{R}",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "first strike")),
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			for _, t := range ctx.ClauseTargets(0) {
				if t.Kind != game.TargetCard || !ctx.IsTargetLegal(t) {
					continue
				}
				if err := (BoostUntilEOT{Target: t.ID, Power: 1, Label: "Abandon Reason — +1/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"first strike"}, Label: "Abandon Reason — first strike"}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
