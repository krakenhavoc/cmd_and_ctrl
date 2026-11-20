package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Expanded Anatomy — Sorcery — Lesson {3} (Strixhaven: School of Mages,
// #5502):
//
//	"Put two +1/+1 counters on target creature. It gains vigilance until
//	 end of turn."
//
// Angelfire Ignition's two counters and keyword grant, with vigilance
// alone. Being a Lesson is its type line, which learn reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6bc5dfbb-98ec-4ff1-8ba2-e2594c53e6d9",
		Name:         "Expanded Anatomy",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (AddCounter{Target: id, Kind: game.CounterPlusOne, N: 2}).Apply(ctx); err != nil {
				return err
			}
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"vigilance"}, Label: "Expanded Anatomy — vigilance"}.Apply(ctx)
		},
	})
}
