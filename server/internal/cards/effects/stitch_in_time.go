package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stitch in Time — Sorcery {1}{U}{R} (Ravnica Remastered, #5948; both faces
// of the reversible card print the same text):
//
//	"Flip a coin. If you win the flip, take an extra turn after this one."
//
// The caster calls the coin and flips once (ADR 0054); a won flip queues one
// extra turn for the caster (CR 500.7). A lost flip does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "59b2a90e-542f-4fb0-b290-ac79dc2892a4",
		Name:         "Stitch in Time",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ctx.Game.FlipCoinForEffect(game.CoinFlipSpec{
				Flipper:  ctx.Controller(),
				Source:   ctx.Source(),
				Question: "Stitch in Time — call the coin flip",
				Coins:    1,
				ThenKey:  stitchInTimeFlipped,
				Asked:    game.NewAskedItem(item),
			})
			return nil
		},
	})
}

// stitchInTimeFlipped is the settled call: a won flip is an extra turn for
// the spell's controller. An on-disk key: never renamed, never reused.
var stitchInTimeFlipped = CoinFlipThen("coin-flip/stitch-in-time", func(ctx *Context, r game.CoinFlipped) error {
	if len(r.Result.Won) == 0 || !r.Result.Won[0] {
		return nil
	}
	return TakeExtraTurn{Player: ctx.Controller(), N: 1}.Apply(ctx)
})
