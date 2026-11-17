package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
// (CR 111.7), so it has nothing to play.
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
			target, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			if IsToken(target) {
				return ExileTarget{Target: id}.Apply(ctx)
			}
			tax := 0
			if target.Owner != ctx.Controller() {
				tax = 2
			}
			return ExileWithPermission{Target: id, WhileExiled: true, CostMore: tax}.Apply(ctx)
		},
	})
}
