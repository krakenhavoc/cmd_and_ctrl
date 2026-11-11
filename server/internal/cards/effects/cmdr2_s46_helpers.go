package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// cmdr2S46PlusOneCounterOnClauseTarget is "put a +1/+1 counter on target
// creature you control" for a triggered ability with one target clause: the
// target is re-checked as the ability resolves (CR 608.2b), and nothing
// happens when it is no longer legal. Bat Colony and Scurry of Squirrels
// share it.
func cmdr2S46PlusOneCounterOnClauseTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	t, ok := ctx.ClauseTarget(0)
	if !ok {
		return nil
	}
	return AddCounter{Target: t.ID, Kind: game.CounterPlusOne, N: 1}.Apply(ctx)
}
