package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Repulsor Bots — Artifact Creature — Robot Hero {4}{U}{U}, 4/4 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 When this creature enters, return up to two other target artifacts
//	 and/or creatures to their owners' hands."
//
// Vronos' up-to-two target group on an enters trigger that excludes the
// source, and its bounce body (b35BounceChosenPermanents).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "facc26d5-bb53-4f91-9b8c-73bf4dd44fcd",
		Name:            "Repulsor Bots",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Repulsor Bots — return up to two other target artifacts and/or creatures to their owners' hands",
				func(g *game.Game, item *game.StackItem) error {
					return b35BounceChosenPermanents(NewContext(g, item))
				}),
				Another(TargetPermanent("up to two other target artifacts and/or creatures", Or(Artifact(), Creature())).WithCount(0, 2))),
		},
	})
}
