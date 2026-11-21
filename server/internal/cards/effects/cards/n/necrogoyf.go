package cardsn

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Necrogoyf — Creature — Lhurgoyf {3}{B}{B}, */4:
//
//	"Necrogoyf's power is equal to the number of creature cards in all
//	 graveyards.
//	 At the beginning of each player's upkeep, that player discards a
//	 card.
//	 Madness {1}{B}{B}"
//
// A power-only layer 7a CDA: the printed toughness 4 stays as the
// base. The discard is the upkeep player's own choice (CR 701.9a), a
// real prompt over their hand, and it is that player's, not the
// controller's, so it reads the upkeep event's actor.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e034a02d-3482-44eb-90dc-c38be104197f",
		Name:         "Necrogoyf",
		Completeness: effects.CompletenessFull,
		Madness:      "{1}{B}{B}",
		Static: []game.StaticAbility{{
			Label:     "Power is equal to the number of creature cards in all graveyards.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7A_CDA,
			AppliesTo: effects.SelfOnly,
			Apply: func(c *game.Characteristic, _ *game.Card, g *game.Game, _ *game.Card) {
				c.Power = effects.B41CreatureCardsInAllGraveyards(g)
			},
		}},
		Triggered: []game.TriggeredAbility{
			effects.AtEachUpkeep("Necrogoyf — that player discards a card", func(g *game.Game, item *game.StackItem) error {
				player := effects.NewContext(g, item).Trigger().Event.Actor
				if player == uuid.Nil {
					return nil
				}
				g.QueueDiscardChoiceForEffect(game.DiscardPrompt{
					Player:   player,
					Source:   item.SourceCardID,
					N:        1,
					Question: "Necrogoyf — discard a card",
				})
				return nil
			}),
		},
	})
}
