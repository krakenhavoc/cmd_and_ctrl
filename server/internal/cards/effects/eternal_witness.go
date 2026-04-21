package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Eternal Witness — 2/1 Human Shaman whose ETB trigger reads:
// "When Eternal Witness enters the battlefield, you may return
// target card from your graveyard to your hand."
//
// S14 sandbox simplifications:
//   - "You may" is treated as "you do." No opt-out prompt in S14
//     (prompt-at-ETB is the same UI gap that blocks Acidic Slime's
//     target-at-ETB; both land with the S19 ability-auto-fire
//     sprint). If the pile is empty, the trigger no-ops instead of
//     fizzling loudly.
//   - Target is auto-picked: the most recently added card in the
//     controller's graveyard (top of pile). Same sandbox posture as
//     Regrowth.
//   - OnETB fires via the direct-call hook in wire.go. Migration to
//     the listener-driven ETB pipeline is deferred to S19 without
//     per-card rewrites — the Spec fields stay stable.
func init() {
	Register(Spec{
		OracleID: "30b24e8e-3b0e-4d8e-90f3-f66eb7c1858c",
		Name:     "Eternal Witness",
		OnETB: func(card *game.Card, ctx *Context) error {
			controller := ctx.PlayerByID(card.Controller)
			if controller == nil || controller.Graveyard.Size() == 0 {
				return nil
			}
			top := controller.Graveyard.Cards[controller.Graveyard.Size()-1].InstanceID
			return ReturnFromGraveyard{Target: top, Dest: game.ZoneHand}.Apply(ctx)
		},
	})
}
