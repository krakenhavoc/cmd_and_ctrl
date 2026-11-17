package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mass Production — Sorcery {5}{W} (The Brothers' War, #5482):
//
//	"Create four 1/1 colorless Soldier artifact creature tokens."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "20150b58-a442-472c-8191-b44581322eeb",
		Name:         "Mass Production",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard(broB05SoldierToken), N: 4}.Apply(ctx)
		},
	})
}
