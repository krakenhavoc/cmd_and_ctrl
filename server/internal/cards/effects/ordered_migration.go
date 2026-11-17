package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Ordered Migration — Sorcery {3}{W}{U} (Invasion, #5616):
//
//	"Domain — Create a 1/1 blue Bird creature token with flying for
//	 each basic land type among lands you control."
//
// The count is DomainCount (CR 305.6's five basic land types, each
// counted once, from the layered lands), read as the spell resolves, and
// the token is the shared "1/1 blue Bird with flying" row. With no basic
// land type among your lands it makes nothing. The token count is not
// declared on the Purpose: no printed number would be true.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee0a59ba-8d34-43d5-b9a5-42015588e3d5",
		Name:         "Ordered Migration",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			n := DomainCount(ctx.Game, ctx.Controller())
			if n == 0 {
				return nil
			}
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 blue Bird with flying"), N: n}.Apply(ctx)
		},
	})
}
