package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// ths_b1_helpers.go — shared bodies for the Theros batch (#5561).
//
// Append-only: add a builder, never change what one means.

// thsPlusOneCountersOnThisIfSameObject is "put N +1/+1 counters on this
// creature" for a heroic trigger (Centaur Battlemaster, Staunch-Hearted
// Warrior): nothing if the source has left the battlefield or come back
// as a new object (CR 400.7).
func thsPlusOneCountersOnThisIfSameObject(n int) Effect {
	return func(g *game.Game, item *game.StackItem) error {
		if sourceIsNewObject(g, item) || !onBattlefield(g, item.SourceCardID) {
			return nil
		}
		return AddCounter{Target: item.SourceCardID, Kind: game.CounterPlusOne, N: n}.Apply(NewContext(g, item))
	}
}
