package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Invoke Justice — Sorcery {1}{W}{W}{W}{W} (Kamigawa: Neon Dynasty):
//
//	"Return target permanent card from your graveyard to the battlefield,
//	 then distribute four +1/+1 counters among any number of creatures
//	 and/or Vehicles target player controls."
//
// Two targets: the card (clause 0) and the player (clause 1). The creatures
// and Vehicles are not targeted, so the controller chooses them and the split
// as the spell resolves, after the card has returned (CR 608.2d), and the
// returned permanent can be one of them when that player is the caster.
// Either target becoming illegal leaves the other half to resolve (CR
// 608.2b). The return puts the card onto the battlefield under its owner's
// control, which is the caster's: it is "your graveyard".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3e5cd60f-8fef-4152-a957-0a803997b3da",
		Name:         "Invoke Justice",
		Completeness: CompletenessFull,
		Targets: Clauses(
			TargetCardInGraveyard("target permanent card from your graveyard", Permanent(), YouOwn()),
			TargetPlayer("target player"),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (ReturnFromGraveyard{Target: t.ID, Dest: game.ZoneBattlefield, Controller: ctx.Controller()}).Apply(ctx); err != nil {
					return err
				}
			}
			return neoDistributeAmongPlayersCreaturesAndVehicles(ctx, 1, 4)
		},
	})
}
