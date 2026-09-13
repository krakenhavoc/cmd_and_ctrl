package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Dark Deal — Sorcery {2}{B} (EDHREC rank 1457):
//
//	"Each player discards all the cards in their hand, then draws
//	 that many cards minus one."
//
// The wheel that shrinks. Wheel of Fortune's body with the count
// read per player: every discard happens before any draw, so
// discard payoffs queue while the effect is still resolving, and
// each player then draws their own hand size less one. A player who
// had nothing draws nothing rather than minus one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c527eb80-ccac-40b8-8377-c31121613128",
		Name:         "Dark Deal",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			players := tablePlayers(ctx)
			draws := make([]int, len(players))
			for i, id := range players {
				n, err := discardWholeHand(ctx.Game, id)
				if err != nil {
					return err
				}
				draws[i] = n - 1
			}
			for i, id := range players {
				if err := (DrawCards{Player: id, N: draws[i]}).Apply(ctx); err != nil {
					return err
				}
			}
			return nil
		},
	})
}
