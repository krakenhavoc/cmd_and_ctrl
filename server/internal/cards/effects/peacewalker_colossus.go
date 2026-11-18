package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Peacewalker Colossus — Artifact — Vehicle {3}, 6/6:
//
//	"{1}{W}: Another target Vehicle you control becomes an artifact
//	 creature until end of turn.
//	 Crew 4"
//
// Chandra, Spark Hunter's Vehicle animation (BecomeCreatureUntilEOT on the
// target) as an activated ability. "Another" is object identity
// (Another, CR 109.1), so the Colossus cannot animate itself with it.
// Crew 4 is the ordinary CR 702.122 crew ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f89f7c7d-6eb3-4335-b9df-0ccaa9c50558",
		Name:         "Peacewalker Colossus",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}{W}: Another target Vehicle you control becomes an artifact creature until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    ManaCost("{1}{W}"),
				Targets: Another(TargetPermanent("another target Vehicle you control", OfSubtype("Vehicle"), YouControl())),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					t, ok := firstLegalTarget(ctx)
					if !ok {
						return nil
					}
					return BecomeCreatureUntilEOT{Target: t.ID, Label: "Peacewalker Colossus — becomes an artifact creature"}.Apply(ctx)
				},
			},
			{
				Label:   "Crew 4",
				Cost:    CrewCost(4),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Peacewalker Colossus"),
			},
		},
	})
}
