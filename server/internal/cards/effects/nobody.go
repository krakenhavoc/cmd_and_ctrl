package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nobody — {1}{U/R}{U/R} Artifact Creature — Human Hero 3/2 (Teenage Mutant
// Ninja Turtles, #4457):
//
//	"When this creature enters, return up to one other target artifact you
//	 control to its owner's hand. Scry 1."
//
// Rescue, Pepper Potts's "up to one other target" enters trigger, narrowed
// to artifacts, followed by scry 1. With no target chosen the trigger still
// resolves and scries; a chosen target that is no longer legal is not
// returned (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eddab60f-2eac-4f20-934c-3d9307f885c3",
		Name:         "Nobody",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Nobody — return up to one other target artifact you control to its owner's hand; scry 1",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := b35BounceChosenPermanents(ctx); err != nil {
						return err
					}
					return Scry{N: 1}.Apply(ctx)
				}),
				Another(TargetPermanent("up to one other target artifact you control", Artifact(), YouControl()).WithCount(0, 1))),
		},
	})
}
