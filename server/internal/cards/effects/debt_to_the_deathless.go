package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Debt to the Deathless — Sorcery {X}{W}{W}{B}{B} (EDHREC rank 2551):
//
//	"Each opponent loses two times X life. You gain life equal to the
//	 life lost this way."
//
// The Orzhov finisher. X is read off the stack item; every opponent
// loses 2X (life loss, not damage — no prevention or doubler sees
// it) and the controller gains the total, 2X per live opponent
// (b21DrainEachOpponentAndGainTheTotal, Gray Merchant's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6130f22d-7901-4f9f-b777-27bb0dacc063",
		Name:         "Debt to the Deathless",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b21DrainEachOpponentAndGainTheTotal(ctx.Game, item, 2*ctx.X())
		},
	})
}
