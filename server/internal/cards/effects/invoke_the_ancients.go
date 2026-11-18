package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invoke the Ancients — Sorcery {1}{G}{G}{G}{G} (Kamigawa: Neon Dynasty):
//
//	"Create two 4/5 green Spirit creature tokens. For each of them, put
//	 your choice of a reach counter, a vigilance counter, or a trample
//	 counter on it."
//
// The tokens land first (a token-doubling replacement makes more of them, and
// each one is asked: "for each of them"), then the controller chooses a
// counter for each in turn. Each choice is its own, so the two Spirits can
// carry the same kind or different ones. The counters are keyword counters the
// engine reads itself (CR 122.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "868b3c65-d753-4f32-b6ca-31e58782ea92",
		Name:         "Invoke the Ancients",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 2},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ctx.Game.CreateTokensKeyedForEffect(game.TokenCreation{
				Controller: item.Controller,
				Source:     ctx.Source(),
				Groups: []game.TokenGroup{{
					Template: ctx.TextToken(TokenCard("4/5 green Spirit")),
					Count:    2,
				}},
			}, ctx.Tail(neoTokenCounterTail))
		},
	})
}
