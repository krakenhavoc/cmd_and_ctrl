package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "27a9e0af-76f0-4910-9f93-2a3617b4e79a",
		Name:         "Spore Burst",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.CreateTokensForEach{
				Template: effects.B11GreenSaprolingToken(),
				Name:     "the basic land types among lands you control",
				Count:    func(ctx *effects.Context) int { return effects.DomainCount(ctx.Game, ctx.Controller()) },
			}.Apply(ctx)
		},
	})
}
