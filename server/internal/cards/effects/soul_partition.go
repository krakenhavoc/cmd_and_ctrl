package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soul Partition — Instant {1}{W} (The Brothers' War):
//
//	"Exile target nonland permanent. For as long as that card remains
//	 exiled, its owner may play it. A spell cast by an opponent this way
//	 costs {2} more to cast."
//
// Zero, Incorporeal Sensate's sentence with the tax only on an OPPONENT's
// cast: the owner of the card is an opponent of Soul Partition's
// controller exactly when the permanent was not theirs, so the {2} rides
// the grant only then (CostMore). A token is exiled and ceases to exist
// (CR 111.7), so it has nothing to play (the grant lands on nothing).
func init() {
	Register(Spec{
		OracleID:     "2cfd5c18-47d0-4549-aacc-2428db21df2f",
		Name:         "Soul Partition",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ExileWithPermission{Target: id, WhileExiled: true, CostMore: soulPartitionTax(ctx, id)}.Apply(ctx)
		},
	})
}

// soulPartitionTax is the {2} on a spell cast through the grant: only an
// OPPONENT of Soul Partition's controller pays it, and that is the owner of
// the exiled card exactly when it was not the controller's own. A token
// ceases to exist in exile (CR 111.7), so the grant over it never lands.
func soulPartitionTax(ctx *Context, id uuid.UUID) int {
	c, ok := ctx.Game.LookupCardForEffect(id)
	if ok && c.Owner != ctx.Controller() {
		return 2
	}
	return 0
}
