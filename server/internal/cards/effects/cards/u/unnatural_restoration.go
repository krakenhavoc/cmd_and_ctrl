package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Unnatural Restoration — Sorcery {1}{G} (Phyrexia: All Will Be One, #5479):
//
//	"Return target permanent card from your graveyard to your hand.
//	 Proliferate."
//
// Regrowth's return with a permanent-card filter, then the player's
// proliferate choice. With its only target gone the spell does not resolve
// at all (CR 608.2b), so there is no proliferate either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "05d1e797-bda2-4411-a546-919493768650",
		Name:         "Unnatural Restoration",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target permanent card in your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard || !ctx.IsTargetLegal(item.Targets[0]) {
				return nil
			}
			if err := effects.ReturnTargetFromGraveyardToHand(item, ctx); err != nil {
				return err
			}
			return effects.Proliferate{}.Apply(ctx)
		},
	})
}
