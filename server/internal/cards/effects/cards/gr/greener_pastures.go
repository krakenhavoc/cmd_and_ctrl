package cardsgr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Greener Pastures — Enchantment {2}{G} (Urza's Saga, S82 OLD batch 732,
// #5624):
//
//	"At the beginning of each player's upkeep, if that player controls more
//	 lands than each other player, the player creates a 1/1 green Saproling
//	 creature token."
//
// "That player" is the one whose upkeep it is, and the token is theirs. The
// "if" is intervening (CR 603.4): that player must control strictly more
// lands than every other player still in the game, as the upkeep begins and
// again as the trigger resolves (a tie, or a land destroyed in response, makes
// it do nothing).
//
// No simplification.
func init() {
	moreLands := func(g *game.Game, player uuid.UUID) bool {
		mine := len(g.LandsControlledForEffect(player))
		for _, p := range g.Seats {
			if p == nil || p.Eliminated || p.ID == player {
				continue
			}
			if len(g.LandsControlledForEffect(p.ID)) >= mine {
				return false
			}
		}
		return mine > 0
	}
	effects.Register(effects.Spec{
		OracleID:     "f9e1346a-af72-4644-b384-8f7b0a35ec31",
		Name:         "Greener Pastures",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Triggered: []game.TriggeredAbility{
			effects.InterveningIf(effects.CondBoard{
				Name: "that player controls more lands than each other player",
				Test: func(ctx *effects.Context) bool {
					return ctx.Item != nil && ctx.Item.Trigger != nil && moreLands(ctx.Game, ctx.Item.Trigger.Event.Actor)
				},
			}, effects.On(game.EventBeginUpkeep, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				return ev.Actor != uuid.Nil && moreLands(g, ev.Actor)
			}, "Greener Pastures — that player creates a 1/1 green Saproling",
				func(g *game.Game, item *game.StackItem) error {
					if item.Trigger == nil {
						return nil
					}
					return effects.CreateToken{Controller: item.Trigger.Event.Actor, Template: effects.TokenCard("1/1 green Saproling"), N: 1}.Apply(effects.NewContext(g, item))
				})),
		},
	})
}
