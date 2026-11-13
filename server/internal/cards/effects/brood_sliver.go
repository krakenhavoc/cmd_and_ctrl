package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Brood Sliver — Creature — Sliver {4}{G}, 3/3 (Legions, #5609):
//
//	"Whenever a Sliver deals combat damage to a player, its controller
//	 may create a 1/1 colorless Sliver creature token."
//
// Cabal Slaver's trigger over Slivers: any Sliver's combat damage to a
// player, whoever controls it, one trigger per Sliver that connects. The
// "may" and the token belong to the Sliver's controller, which the
// damage event carries in Actor as the damage is dealt (Edric,
// Spymaster of Trest's shape), so they are still that player's when the
// Sliver has died to the same combat damage. The trigger is the Brood
// Sliver's, so it goes on the stack under the Brood Sliver's controller.
//
// No simplification.
func init() {
	t := On(game.EventDealDamage, aCreatureOfTypeDealtCombatDamageToAPlayer("Sliver"),
		"Brood Sliver — its controller may create a 1/1 colorless Sliver creature token",
		func(g *game.Game, item *game.StackItem) error {
			ctx := NewContext(g, item)
			owner := ctx.Trigger().Event.Actor
			if g.PlayerByIDForEffect(owner) == nil {
				return nil
			}
			return CreateToken{Controller: owner, Template: TokenCard("1/1 colorless Sliver"), N: 1}.Apply(ctx)
		})
	t.OptionalPrompt = &game.TriggerOptionalPrompt{
		Question: "Brood Sliver — create a 1/1 colorless Sliver creature token?",
		Chooser: func(ev game.Event, _ *game.Card, _ *game.Game) uuid.UUID {
			return ev.Actor
		},
	}
	Register(Spec{
		OracleID:     "6b1d178f-9713-4dcf-920b-ddf2c94cc427",
		Name:         "Brood Sliver",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{t},
	})
}
