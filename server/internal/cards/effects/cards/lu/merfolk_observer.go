package cardslu

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Merfolk Observer — Creature — Merfolk Rogue {1}{U}, 2/1 (Rise of the
// Eldrazi, S82 OLD batch 484, #5581):
//
//	"When this creature enters, look at the top card of target player's
//	 library."
//
// Only the controller sees the card (Draugr Thought-Thief's look, without its
// graveyard option); nothing moves.
//
// No simplification.
func init() {
	enters := effects.WhenThisEnters("Merfolk Observer — look at the top card of target player's library",
		func(g *game.Game, item *game.StackItem) error {
			ctx := effects.NewContext(g, item)
			for _, t := range ctx.LegalTargets() {
				if t.Kind == game.TargetPlayer && t.ID != uuid.Nil {
					g.LookAtTopOfPlayersLibraryForEffect(item.Controller, t.ID, 1)
					return nil
				}
			}
			return nil
		})
	enters.Targets = effects.TargetPlayer("target player")
	effects.Register(effects.Spec{
		OracleID:     "0b5d49ce-7e02-4de6-8a65-bfc88ac0cf72",
		Name:         "Merfolk Observer",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{enters},
	})
}
