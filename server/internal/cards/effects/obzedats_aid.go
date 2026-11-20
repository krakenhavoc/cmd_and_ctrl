package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obzedat's Aid — Sorcery {3}{W}{B}:
//
//	"Return target permanent card from your graveyard to the
//	 battlefield."
//
// Profound Journey without rebound: any permanent card in its controller's
// graveyard returns to the battlefield.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f3be91ce-e99b-4a51-a0f7-a000fd9d0e14",
		Name:         "Obzedat's Aid",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target permanent card in your graveyard", YouOwn(), Permanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return returnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}
