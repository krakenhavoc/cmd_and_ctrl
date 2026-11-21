package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Restore with Empathy — Instant {2}{G} (Reality Fracture, tracker #2795):
//
//	"Return target permanent card from your graveyard to your hand. You
//	 gain 4 life."
//
// A spell whose one target has gone does nothing at all (CR 608.2b), so
// the life is not gained either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "cb1e697f-6689-4f0b-8725-61d51432b8dd",
		Name:         "Restore with Empathy",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCardInGraveyard("target permanent card in your graveyard", effects.YouOwn(), effects.Permanent()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(effects.LegalTargetCardIDs(ctx)) == 0 {
				return nil
			}
			if err := effects.ReturnTargetGraveyardCardToHand(ctx.Game, item); err != nil {
				return err
			}
			return effects.GainLife{Player: ctx.Controller(), Amount: 4}.Apply(ctx)
		},
	})
}
