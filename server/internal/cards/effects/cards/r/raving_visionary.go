package cardsr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Raving Visionary — Creature — Merfolk Wizard {1}{U}, 1/1 (Modern
// Horizons 2, #5501):
//
//	"{U}, {T}: Draw a card, then discard a card.
//	 Delirium — {2}{U}, {T}: Draw a card. Activate only if there are four
//	 or more card types among cards in your graveyard."
//
// Strix Lookout's loot at {U}, and a draw gated by Balustrade Wurm's
// delirium Condition (b16CardTypesInGraveyard), checked as the ability
// is activated (CR 602.5b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "83157972-3a0b-40e2-a80a-7bfb2a8627ee",
		Name:         "Raving Visionary",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "{U}, {T}: Draw a card, then discard a card.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{U}"), effects.TapCost()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.B16DrawThenDiscard(g, item, 1)
				},
			},
			{
				Label:   "Delirium — {2}{U}, {T}: Draw a card. Activate only if there are four or more card types among cards in your graveyard.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}{U}"), effects.TapCost()),
				Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
					return effects.B16CardTypesInGraveyard(g, controller) >= 4
				},
				Effect: effects.EbDrawACard,
			},
		},
	})
}
