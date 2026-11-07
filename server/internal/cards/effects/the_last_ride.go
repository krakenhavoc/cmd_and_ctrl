package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Last Ride — Legendary Artifact — Vehicle {B}, 13/13 (Aetherdrift,
// #4464):
//
//	"The Last Ride gets -X/-X, where X is your life total.
//	 {2}{B}, Pay 2 life: Draw a card.
//	 Crew 2"
//
// The -X/-X is a layer 7c modification (CR 613.4c) read off Serra
// Avatar's yourLifeTotal, so it applies whether or not the Vehicle is a
// creature right now. DependsOnLifeTotal tells the layer cache to
// recompute as the life total moves, as Serra Ascendant's does. A
// negative life total gives X = 0, not a bonus (CR 107.1b). The draw is
// Dr. Madison Li's draw row with a life payment (CR 119.4) in place of
// energy. Crew is the engine's (CR 702.122a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3ef03540-b538-466d-aa3c-da6750117ba4",
		Name:         "The Last Ride",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{{
			Label:              "The Last Ride gets -X/-X, where X is your life total.",
			Layer:              game.Layer7PT,
			SubLayer:           game.SubLayer7C_Modify,
			AppliesTo:          selfOnly,
			DependsOnLifeTotal: true,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, source *game.Card) {
				x := max(yourLifeTotal(g, source), 0)
				c.Power -= x
				c.Toughness -= x
			},
		}},
		Activated: []ActivatedAbility{
			{
				Label:   "{2}{B}, Pay 2 life: Draw a card.",
				Cost:    Plus(ManaCost("{2}{B}"), PayLife(2)),
				Purpose: game.Purpose{Answers: game.AnswerValue, Draws: 1},
				Effect:  Do(DrawCards{N: 1}),
			},
			{
				Label:   "Crew 2",
				Cost:    CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  CrewEffect("The Last Ride"),
			},
		},
	})
}
