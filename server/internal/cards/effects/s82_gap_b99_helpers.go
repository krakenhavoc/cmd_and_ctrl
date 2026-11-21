package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_gap_b99_helpers.go — shared bodies for GAP batch 99 (#5948).

// gapB99CounterOnThis is "put a +1/+1 counter on <this permanent>" as a
// trigger body: nothing happens if the source has left the battlefield
// (Omarthis).
func gapB99CounterOnThis(g *game.Game, item *game.StackItem) error {
	if !onBattlefield(g, item.SourceCardID) {
		return nil
	}
	return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: 1}.Apply(NewContext(g, item))
}
