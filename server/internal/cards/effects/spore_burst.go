package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spore Burst — Sorcery {3}{G} (Conflux, #5587):
//
//	"Domain — Create a 1/1 green Saproling creature token for each basic
//	 land type among lands you control."
//
// The count is the shared DomainCount (domain.go, CR 305.6), read as the
// spell resolves; the token is the table's 1/1 green Saproling.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "27a9e0af-76f0-4910-9f93-2a3617b4e79a",
		Name:         "Spore Burst",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateTokensForEach{
				Template: b11GreenSaprolingToken(),
				Name:     "the basic land types among lands you control",
				Count:    func(ctx *Context) int { return DomainCount(ctx.Game, ctx.Controller()) },
			}.Apply(ctx)
		},
	})
}
