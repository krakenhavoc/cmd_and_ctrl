package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "568432d1-c7fb-4da6-922a-22d4fa26284e",
		Name:         "Capenna Express",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Sacrifice a Treasure: This Vehicle becomes an artifact creature until end of turn.",
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Cost:    effects.SacrificeN(1, "a Treasure", effects.Subtype("Treasure")),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.BecomeCreatureUntilEOT{Label: "Capenna Express — becomes an artifact creature"}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "Crew 3",
				Cost:    effects.CrewCost(3),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Capenna Express"),
			},
		},
	})
}
