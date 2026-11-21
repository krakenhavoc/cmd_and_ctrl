package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "f3be91ce-e99b-4a51-a0f7-a000fd9d0e14",
		Name:         "Obzedat's Aid",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target permanent card in your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			return effects.ReturnFirstLegalGraveyardTargetToBattlefield(ctx.Game, item)
		},
	})
}
