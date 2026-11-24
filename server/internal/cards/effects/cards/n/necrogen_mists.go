package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necrogen Mists — Enchantment {2}{B} (Secret Lair Drop, #5462):
//
//	"At the beginning of each player's upkeep, that player discards a
//	 card."
//
// Bottomless Pit with a choice: the player whose upkeep it is, read off the
// event, picks the card to discard (Oppression's prompt); an empty hand
// discards nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "db25d48e-e5eb-4364-9d7a-f9757ee4ccbe",
		Name:         "Necrogen Mists",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Necrogen Mists — that player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TriggeringActor(item)
					if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
						return nil
					}
					g.QueueDiscardChoiceForEffect(game.DiscardPrompt{Player: player, Source: item.SourceCardID, N: 1})
					return nil
				}),
		},
	})
}
