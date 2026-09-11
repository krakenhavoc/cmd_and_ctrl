package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wheel of Fortune — Sorcery {2}{R} (EDHREC rank 569):
//
//	"Each player discards their hand, then draws seven cards."
//
// The original wheel. Every discard happens before any draw, and
// each discard emits its own event, so discard payoffs queue while
// the spell is still resolving — Windfall's shape with a fixed seven
// instead of "the most anyone discarded". A player with an empty
// hand still draws seven.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "a8abd966-de7b-46a3-8ac7-8747ab35653a",
		Name:     "Wheel of Fortune",
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			players := tablePlayers(ctx)
			for _, id := range players {
				if _, err := discardWholeHand(ctx.Game, id); err != nil {
					return err
				}
			}
			for _, id := range players {
				if err := (DrawCards{Player: id, N: 7}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
