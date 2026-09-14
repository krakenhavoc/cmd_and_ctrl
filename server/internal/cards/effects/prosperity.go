package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Prosperity — Sorcery {X}{U} (EDHREC rank 3634):
//
//	"Each player draws X cards."
//
// The group hug X-draw. X is the announce-time value the cost engine
// charged; every seated player draws that many, APNAP from the
// active seat, each an ordinary draw so "whenever an opponent draws"
// payoffs see every card. X of zero draws nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c586312d-d04a-4bfb-bbb2-b41186ca178e",
		Name:         "Prosperity",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return b05EachPlayerDraws(ctx.Game, item, ctx.X())
		},
	})
}
