package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Scaled Herbalist — Creature — Lizard Druid {1}{G}, 1/3 (Adventures in
// the Forgotten Realms):
//
//	"{T}: You may put a land card from your hand onto the battlefield."
//
// Budoka Gardener's land drop (MayPutALandFromHand): optional, from the
// hand only, untapped, and not a land play.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6fd9ece7-366f-45c7-a8ea-edb4ca637a06",
		Name:         "Scaled Herbalist",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{T}: You may put a land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.MayPutALandFromHand("Scaled Herbalist").Apply(effects.NewContext(g, item))
			},
		}},
	})
}
