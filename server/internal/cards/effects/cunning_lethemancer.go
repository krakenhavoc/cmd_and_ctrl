package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Cunning Lethemancer — Creature — Human Wizard {2}{B}, 2/2 (Shards of
// Alara, #5588):
//
//	"At the beginning of your upkeep, each player discards a card."
//
// Liliana of the Veil's +1 on an upkeep trigger: every player in the
// game, its controller included, is asked for one card from their own
// hand, in turn order from the active player, and a player with an empty
// hand discards nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d696aba9-ef3e-4f81-821a-c1df51431b8a",
		Name:         "Cunning Lethemancer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Cunning Lethemancer — each player discards a card", func(g *game.Game, item *game.StackItem) error {
				g.EachPlayerDiscardsForEffect(uuid.Nil, game.DiscardPrompt{
					Source: item.SourceCardID, N: 1, Question: "Cunning Lethemancer — discard a card",
				})
				return nil
			}),
		},
	})
}
