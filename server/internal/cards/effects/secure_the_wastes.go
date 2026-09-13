package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Secure the Wastes — Instant {X}{W} (EDHREC rank 1247):
//
//	"Create X 1/1 white Warrior creature tokens."
//
// An instant-speed army for X. The X is announced with the cast and
// read back with ctx.X(); zero tokens for X=0 is a legal, pointless
// cast, as printed.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b2347910-d6c6-4681-8316-7ef27056485c",
		Name:         "Secure the Wastes",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: b11WhiteWarriorToken(), N: ctx.X()}.Apply(ctx)
		},
	})
}
