package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Soaring Show-Off — Creature — Bird Warrior {2}{U}, 2/2 (Battlebond,
// #5529):
//
//	"Flying
//	 When this creature enters, each player draws a card."
//
// Flying rides PrintedKeywords. The enters trigger is the shared "each
// player draws N" (b05EachPlayerDraws): every player still in the game
// draws one card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "5ca0f4f3-6133-4f53-9124-f63942796a07",
		Name:            "Soaring Show-Off",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Soaring Show-Off — each player draws a card", func(g *game.Game, item *game.StackItem) error {
				return effects.B05EachPlayerDraws(g, item, 1)
			}),
		},
	})
}
