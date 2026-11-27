package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tainted Aether — Enchantment {2}{B}{B} (Urza's Saga, S82 OLD batch 730,
// #5624):
//
//	"Whenever a creature enters, its controller sacrifices a creature or land
//	 of their choice."
//
// Any creature entering, anyone's, including the creature that triggers it.
// "Its controller" is read as the trigger resolves, from the entered
// permanent as it is now or as it last existed there (CR 608.2h); that player
// chooses a creature or a land they control, the newcomer being a legal
// choice, and sacrifices it. With neither, nothing happens.
//
// No simplification.
func init() {
	const label = "Tainted Aether — its controller sacrifices a creature or land"
	effects.Register(effects.Spec{
		OracleID:     "a61ceda1-5993-479e-945f-15753eeb7049",
		Name:         "Tainted Aether",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature()
			}, label, func(g *game.Game, item *game.StackItem) error {
				info, ok := effects.NewContext(g, item).TriggeringPermanent()
				if !ok {
					return nil
				}
				g.PlayerSacrificesForEffect(item.SourceCardID, info.Controller,
					effects.SacrificeSpec("a creature or land", effects.Or(effects.Creature(), effects.Land())), label)
				return nil
			}),
		},
	})
}
