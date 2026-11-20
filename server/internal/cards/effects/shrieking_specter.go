package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shrieking Specter — Creature — Specter {5}{B}, 2/2 (Starter 1999,
// #5620):
//
//	"Flying
//	 Whenever this creature attacks, defending player discards a card."
//
// Witch-king, Bringer of Ruin's attack trigger with Abyssal
// Nightstalker's payoff: the defending player is the player attacked, or
// the controller of the planeswalker or battle attacked
// (defendingPlayerOfAttack), and they choose the card they discard, so
// the payoff is the discard prompt (CR 701.9a). A defender who has left
// the game discards nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f6adfb5f-a21c-4e86-bd16-a9c1e355af34",
		Name:            "Shrieking Specter",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			WheneverThisAttacks("Shrieking Specter — defending player discards a card",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					defender := defendingPlayerOfAttack(g, item.Trigger.Event)
					if defender == uuid.Nil {
						return nil
					}
					return defendingPlayerDiscardsACard(g, item, defender)
				}),
		},
	})
}
