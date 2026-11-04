package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Second Harvest — Instant {2}{G}{G} (EDHREC top 1,000, #3654):
//
//	"For each token you control, create a token that's a copy of that
//	 permanent."
//
// The tokens you control are fixed when the spell starts to resolve, so
// the copies it makes are not themselves copied. The copies are one
// creation instruction (#4594, CreateTokenCopiesTogether): they are
// created at the same time (CR 608.2f), so they enter together, and a
// token-doubling replacement doubles every one of them.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61a6341e-0d36-44b6-8026-9c87c8ef61cb",
		Name:         "Second Harvest",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			var ids []uuid.UUID
			for _, c := range ctx.Game.BattlefieldCardsForEffect() {
				if c.Controller == item.Controller && c.IsToken() {
					ids = append(ids, c.InstanceID)
				}
			}
			return CreateTokenCopiesTogether{Controller: item.Controller, Copies: ids}.Apply(ctx)
		},
	})
}
