package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Synapse Sliver — Creature — Sliver {4}{U}, 3/3 (Commander Legends: Battle
// for Baldur's Gate):
//
//	"Whenever a Sliver deals combat damage to a player, its controller may
//	 draw a card."
//
// Brood Sliver's trigger: any Sliver at the table connecting, the "may" put
// to the Sliver's controller (the event's actor), who then draws.
//
// No simplification.
func init() {
	t := On(game.EventDealDamage, aCreatureOfTypeDealtCombatDamageToAPlayer("Sliver"),
		"Synapse Sliver — its controller may draw a card",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			owner := ctx.Trigger().Event.Actor
			if g.PlayerByIDForEffect(owner) == nil {
				return nil
			}
			return DrawCards{Player: owner, N: 1}.Apply(ctx)
		})
	t.OptionalPrompt = &game.TriggerOptionalPrompt{
		Question: "Synapse Sliver — draw a card?",
		Chooser: func(ev game.Event, _ *game.Card, _ *game.Game) uuid.UUID {
			return ev.Actor
		},
	}
	Register(Spec{
		OracleID:     "b76f123d-7b78-42a1-a66e-0dee6f773687",
		Name:         "Synapse Sliver",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
