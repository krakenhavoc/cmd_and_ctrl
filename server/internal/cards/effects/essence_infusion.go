package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Essence Infusion — Sorcery {1}{B} (Strixhaven: School of Mages, #5502):
//
//	"Put two +1/+1 counters on target creature. It gains lifelink until
//	 end of turn."
//
// Angelfire Ignition's two counters and keyword grant, with lifelink
// alone.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "60f412bb-27e9-4883-8862-96eae96dfee3",
		Name:         "Essence Infusion",
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
			return GrantKeywordUntilEOT{Target: id, Keywords: []string{"lifelink"}, Label: "Essence Infusion — lifelink"}.Apply(ctx)
		},
	})
}
