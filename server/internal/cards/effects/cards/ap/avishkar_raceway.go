package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Avishkar Raceway — Land:
//
//	"Start your engines!
//	 {T}: Add {C}.
//	 Max speed — {3}, {T}, Discard a card: Draw a card."
//
// ADR 0138 (#2122). The rummage is an ordinary activated ability with a
// discard cost (paid at activation, CR 602.2b), activatable only while
// its controller has max speed (CR 702.178a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "e2c35551-1ba5-4424-baf9-821b49bbcc8c",
		Name:            "Avishkar Raceway",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{C}", Label: "{T}: Add {C}."},
		},
		Activated: []effects.ActivatedAbility{
			effects.MaxSpeedActivated(effects.ActivatedAbility{
				Label:   "Max speed — {3}, {T}, Discard a card: Draw a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost(), effects.DiscardACard()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
				},
			}),
		},
	})
}
