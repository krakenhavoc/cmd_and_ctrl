package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Soldier On — Instant {2}{W} (Marvel Super Heroes Commander, #4455):
// "Put two +1/+1 counters on target creature. Untap it." Two steps on
// the one target: the counters, then the untap.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f522e161-f249-4add-8326-b6ab0de8ba57",
		Name:         "Soldier On",
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
			return UntapTarget{Target: id}.Apply(ctx)
		},
	})
}
