package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deploy to the Front — Sorcery {5}{W}{W} (Commander 2014, #5553):
//
//	"Create X 1/1 white Soldier creature tokens, where X is the number of
//	 creatures on the battlefield."
//
// X is counted once as the spell resolves, before any Soldier is made, and
// it counts every creature on the battlefield, whoever controls it (Chain
// Reaction's count). An empty battlefield makes nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "135cb4af-f4b6-455b-8fea-84fcef78ee03",
		Name:         "Deploy to the Front",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateTokensForEach{
				Template: effects.TokenCard("1/1 white Soldier"),
				Name:     "creatures on the battlefield",
				Count:    func(ctx *effects.Context) int { return len(effects.MatchingBattlefield(ctx, effects.Creature())) },
			}.Apply(ctx)
		},
	})
}
