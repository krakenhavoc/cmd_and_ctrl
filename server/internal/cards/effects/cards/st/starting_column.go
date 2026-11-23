package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Starting Column — Artifact {3}:
//
//	"Start your engines!
//	 {T}: Add one mana of any color.
//	 Max speed — {T}, Sacrifice this artifact: Draw two cards, then
//	 discard a card."
//
// ADR 0138 (#2122). The loot is an ordinary activated ability,
// activatable only while its controller has max speed (CR 702.178a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d608a3fa-faee-44a4-9ab6-be701d3b7a49",
		Name:            "Starting Column",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		ManaAbilities: []effects.ManaAbility{
			{Cost: effects.ManaAbilityCost{Tap: true}, Produced: "{W|U|B|R|G}", Label: "{T}: Add one mana of any color."},
		},
		Activated: []effects.ActivatedAbility{
			effects.MaxSpeedActivated(effects.ActivatedAbility{
				Label:   "Max speed — {T}, Sacrifice this artifact: Draw two cards, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.TapCost(), effects.SacrificeThis()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B16DrawThenDiscard(g, item, 2)
				},
			}),
		},
	})
}
