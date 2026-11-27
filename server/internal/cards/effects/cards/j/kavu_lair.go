package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kavu Lair — Enchantment {2}{G} (Invasion, S82 #5616):
//
//	"Whenever a creature with power 4 or greater enters, its controller draws a
//	 card."
//
// Any creature, anyone's: the power is read as it enters (counters included),
// and "its controller" is read from the entering permanent when the trigger
// resolves, as it last existed if it has left since (CR 608.2h).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4dd0594d-5b11-4350-a926-294a67d9fc3b",
		Name:         "Kavu Lair",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := g.LookupCardForEffect(ev.CardID)
				return ok && c.IsCreature() && c.CurrentPower() >= 4
			}, "Kavu Lair — its controller draws a card", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				info, ok := ctx.TriggeringPermanent()
				if !ok {
					return nil
				}
				return effects.DrawCards{N: 1, Player: info.Controller}.Apply(ctx)
			}),
		},
	})
}
