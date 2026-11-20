package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Glimmerpoint Stag — Creature — Elk {2}{W}{W}, 3/3 (Scars of Mirrodin,
// #5577):
//
//	"Vigilance
//	 When this creature enters, exile another target permanent. Return
//	 that card to the battlefield under its owner's control at the
//	 beginning of the next end step."
//
// The shared flicker body (exileTargetsThenScheduleReturn) on a required
// target, any permanent but the Stag itself: one exile, then a delayed
// trigger at the next end step (CR 603.7a) that returns what reached
// exile under its owner's control. A token exiled this way does not come
// back (CR 111.8).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9620c8a2-63a8-4151-b4fd-a15db191a7b0",
		Name:            "Glimmerpoint Stag",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Glimmerpoint Stag — exile another target permanent until the next end step",
					Do(Instruction(func(ctx *Context) error {
						return exileTargetsThenScheduleReturn(ctx, "Glimmerpoint Stag — return the exiled card")
					}))),
				Another(TargetPermanent("another target permanent"))),
		},
	})
}
