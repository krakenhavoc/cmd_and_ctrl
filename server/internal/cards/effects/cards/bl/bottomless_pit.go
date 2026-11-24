package cardsbl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Bottomless Pit — Enchantment {1}{B}{B} (Secret Lair Drop, #5462):
//
//	"At the beginning of each player's upkeep, that player discards a card
//	 at random."
//
// Rite of the Raging Storm's every-upkeep trigger: the player whose upkeep
// it is, read off the event, discards one card from their hand at random
// (the keyed random stream, Black Cat's discard); an empty hand discards
// nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "91e6fb47-59e4-4616-b8dd-3a7e30070074",
		Name:         "Bottomless Pit",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Bottomless Pit — that player discards a card at random",
				func(g *game.Game, item *game.StackItem) error {
					player := effects.TriggeringActor(item)
					if player == uuid.Nil || g.PlayerByIDForEffect(player) == nil {
						return nil
					}
					return effects.DiscardCards{Player: player, N: 1}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
