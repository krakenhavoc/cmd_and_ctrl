package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "9620c8a2-63a8-4151-b4fd-a15db191a7b0",
		Name:            "Glimmerpoint Stag",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.WhenThisEnters("Glimmerpoint Stag — exile another target permanent until the next end step",
					effects.Do(effects.Instruction(func(ctx *effects.Context) error {
						return effects.ExileTargetsThenScheduleReturn(ctx, "Glimmerpoint Stag — return the exiled card")
					}))),
				effects.Another(effects.TargetPermanent("another target permanent"))),
		},
	})
}
