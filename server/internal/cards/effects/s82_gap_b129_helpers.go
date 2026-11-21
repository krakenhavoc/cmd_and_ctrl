package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b129_helpers.go — GAP batch 129 (S82 #5948): the one body two of the
// batch's cards share.

// gapB129PutTwoCountersOnTarget is "put two +1/+1 counters on target
// creature you control": the target announced with the ability, checked again
// as it resolves (CR 608.2b), the counters put on it if it is still legal.
func GapB129PutTwoCountersOnTarget(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	targets := ctx.LegalTargets()
	if len(targets) == 0 {
		return nil
	}
	return AddCounter{Target: targets[0].ID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
}
