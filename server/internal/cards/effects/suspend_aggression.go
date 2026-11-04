package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Suspend Aggression — Instant {1}{R}{W} (Secrets of Strixhaven, #4456):
//
//	"Exile target nonland permanent and the top card of your library. For
//	 each of those cards, its owner may play it until the end of their next
//	 turn."
//
// One resolution, one target. If the target is illegal as the spell resolves
// the whole spell does nothing (CR 608.2b) and the library is not touched.
// Otherwise the permanent is exiled first and then the top card of the
// caster's library. Each card's OWNER, not the caster, gets a permission to
// play it (a land or a spell) until the end of that player's next turn: a
// stolen permanent goes back to its owner's hands to play, and a window that
// is the owner's own, stamped now (Game.UntilEndOfYourNextTurnDuration), so
// it is the next turn that player has even when this is their own turn. A
// token that is exiled ceases to exist and has nothing to play.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "883267f0-be7e-43c3-b7f8-82981f6d8c23",
		Name:         "Suspend Aggression",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			target, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			if IsToken(target) {
				if err := (ExileTarget{Target: id}).Apply(ctx); err != nil {
					return err
				}
			} else if err := ctx.Game.ExileCardWithPermissionForEffect(id, game.CastPermission{
				Duration: ctx.Game.UntilEndOfYourNextTurnDuration(target.Owner),
			}); err != nil {
				return err
			}
			window := ctx.Game.UntilEndOfYourNextTurnDuration(ctx.Controller())
			return ExileTopWithPermission{
				From:    ctx.Controller(),
				GrantTo: ctx.Controller(),
				N:       1,
				Window:  &window,
			}.Apply(ctx)
		},
	})
}
