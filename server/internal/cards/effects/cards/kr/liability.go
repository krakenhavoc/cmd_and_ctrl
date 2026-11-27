package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Liability — Enchantment {1}{B}{B} (Mercadian Masques, #5619):
//
//	"Whenever a nontoken permanent is put into a player's graveyard from
//	 the battlefield, that player loses 1 life."
//
// Soulcatchers' Aerie's graveyard-arrival trigger over every permanent:
// "that player" is the owner of the card, since a graveyard is its
// owner's. A token is not counted, and neither is a permanent that went
// to exile or the command zone instead.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "65f1bebb-6441-4189-b6de-993ec9d24a59",
		Name:         "Liability",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				if ev.Kind != game.EventLTB || ev.NewZone != game.ZoneGraveyard {
					return false
				}
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && !c.IsToken()
			}, "Liability — that player loses 1 life", func(g *game.Game, item *game.StackItem) error {
				c, ok := g.LookupCardForEffect(item.Trigger.Event.CardID)
				if !ok {
					return nil
				}
				return g.ChangePlayerLifeForEffect(item.SourceCardID, c.Owner, -1)
			}),
		},
	})
}
