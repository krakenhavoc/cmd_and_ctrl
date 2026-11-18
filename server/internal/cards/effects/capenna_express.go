package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Capenna Express — Artifact — Vehicle {3}{G}, 6/6 (Streets of New
// Capenna):
//
//	"Sacrifice a Treasure: This Vehicle becomes an artifact creature
//	 until end of turn.
//	 Crew 3"
//
// The first ability is a second way to animate it: the sacrifice is the
// cost, and the effect is the same artifact-creature change as crew's
// (CR 702.122), without tapping anyone or recording a crew payment.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "568432d1-c7fb-4da6-922a-22d4fa26284e",
		Name:         "Capenna Express",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "Sacrifice a Treasure: This Vehicle becomes an artifact creature until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    SacrificeN(1, "a Treasure", Subtype("Treasure")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return BecomeCreatureUntilEOT{Label: "Capenna Express — becomes an artifact creature"}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "Crew 3",
				Cost:    CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("Capenna Express"),
			},
		},
	})
}
