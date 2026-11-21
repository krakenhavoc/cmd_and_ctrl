package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rise Again — Sorcery {4}{B} (Core Set 2021, #5511):
//
//	"Return target creature card from your graveyard to the battlefield."
//
// Zombify's reanimation: your graveyard only, so the owner and the new
// controller are the same player.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4600e968-fe3a-446a-a306-bd10e6763e0d",
		Name:         "Rise Again",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreatureInYourGraveyard(),
		Purpose:      effects.ForTargets(effects.Reanimates(0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			_, _ = effects.ReanimateSingleTarget(ctx, ctx.Controller())
			return nil
		},
	})
}
