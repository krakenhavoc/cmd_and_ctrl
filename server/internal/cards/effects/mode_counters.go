package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// PlusOneCountersOnTheModesTarget is "put n +1/+1 counters on target
// <thing>" as a modal bullet's body: The Fantastic Four's and Quandrix
// Command's "put two +1/+1 counters on target creature" bullets differ
// only in the card around them.
func PlusOneCountersOnTheModesTarget(n int) func(item *game.StackItem, ctx *Context, occ int) error {
	return func(_ *game.StackItem, ctx *Context, occ int) error {
		t, ok := ModeTarget(ctx, occ)
		if !ok {
			return nil
		}
		return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: n}.Apply(ctx)
	}
}
