package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Flickerwisp — Creature — Elemental {1}{W}{W}, 3/1 (Eventide, #5590):
//
//	"Flying
//	 When this creature enters, exile another target permanent. Return that
//	 card to the battlefield under its owner's control at the beginning of
//	 the next end step."
//
// Glimmerpoint Stag's flicker: the exiled card comes back under its
// owner's control as a new object at the beginning of the next end step.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b23a3d30-6b8e-4aad-890f-db0c3af43ace",
		Name:            "Flickerwisp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Flickerwisp — exile another target permanent until the next end step",
					Do(Instruction(func(ctx *Context) error {
						return exileTargetsThenScheduleReturn(ctx, "Flickerwisp — return the exiled card")
					}))),
				Another(TargetPermanent("another target permanent"))),
		},
	})
}
