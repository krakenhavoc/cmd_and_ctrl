package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Mobilizer Mech — {1}{U} Artifact — Vehicle, 3/4 (Kamigawa: Neon
// Dynasty, #5671):
//
//	"Flying
//	 Whenever this Vehicle becomes crewed, up to one other target Vehicle
//	 you control becomes an artifact creature until end of turn.
//	 Crew 3"
//
// "Becomes crewed" is WhenBecomesCrewed (CR 702.122e): the trigger goes
// on the stack each time a crew ability of the Mech resolves, so the
// Mech is already a creature when the other Vehicle is chosen. "Up to
// one other": the trigger may be put on the stack with no target, and
// the Mech itself is never a legal one. The animation is the crew
// effect's own, on the chosen Vehicle, which does not count as crewing
// it (nobody crewed it, and its own "becomes crewed" does not trigger).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "3c1a2a2d-656f-4021-b046-59e14383a8c8",
		Name:            "Mobilizer Mech",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  CrewEffect("Mobilizer Mech"),
		}},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenBecomesCrewed("Mobilizer Mech — up to one other target Vehicle you control becomes an artifact creature until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return BecomeCreatureUntilEOT{Target: t.ID, Label: "Mobilizer Mech — becomes an artifact creature"}.Apply(ctx)
						}
					}
					return nil
				}),
				Another(TargetPermanent("up to one other target Vehicle you control", OfSubtype("Vehicle"), YouControl()).WithCount(0, 1))),
		},
	})
}
