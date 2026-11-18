package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "6fd9ece7-366f-45c7-a8ea-edb4ca637a06",
		Name:         "Scaled Herbalist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: You may put a land card from your hand onto the battlefield.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    TapCost(),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return MayPutALandFromHand("Scaled Herbalist").Apply(NewContext(g, item))
			},
		}},
	})
}
