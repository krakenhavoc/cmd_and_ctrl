package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "cb1e697f-6689-4f0b-8725-61d51432b8dd",
		Name:         "Restore with Empathy",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target permanent card in your graveyard", YouOwn(), Permanent()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(legalTargetCardIDs(ctx)) == 0 {
				return nil
			}
			if err := returnTargetGraveyardCardToHand(ctx.Game, item); err != nil {
				return err
			}
			return GainLife{Player: ctx.Controller(), Amount: 4}.Apply(ctx)
		},
	})
}
