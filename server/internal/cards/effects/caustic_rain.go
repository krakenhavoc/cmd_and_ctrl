package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Caustic Rain — Sorcery {2}{B}{B} (Guildpact, #5599):
//
//	"Exile target land."
//
// ExileFirstTarget over any land, checked again as the spell resolves
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8df1cbbc-87fc-4948-b224-9175b9f396d9",
		Name:         "Caustic Rain",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return ExileFirstTarget(ctx.Game, item)
		},
	})
}
