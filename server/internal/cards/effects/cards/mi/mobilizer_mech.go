package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:        "3c1a2a2d-656f-4021-b046-59e14383a8c8",
		Name:            "Mobilizer Mech",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 3",
			Cost:    effects.CrewCost(3),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Mobilizer Mech"),
		}},
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.WhenBecomesCrewed("Mobilizer Mech — up to one other target Vehicle you control becomes an artifact creature until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetCard {
							return effects.BecomeCreatureUntilEOT{Target: t.ID, Label: "Mobilizer Mech — becomes an artifact creature"}.Apply(ctx)
						}
					}
					return nil
				}),
				effects.Another(effects.TargetPermanent("up to one other target Vehicle you control", effects.OfSubtype("Vehicle"), effects.YouControl()).WithCount(0, 1))),
		},
	})
}
