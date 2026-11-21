package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Laboratory Drudge — Creature — Zombie Horror {3}{U}, 3/4 (Commander Starter
// Decks):
//
//	"At the beginning of each end step, draw a card if you've cast a spell
//	 from a graveyard or activated an ability of a card in a graveyard this
//	 turn."
//
// Every end step, not just its controller's. The "if" follows the effect, so
// it is read as the ability resolves and not as an intervening if (CR 603.4).
// It reads the turn's event log: an EventCast of yours cast from a graveyard
// (its OldZone), or an EventActivateAbility of yours whose ActivatedFrom is a
// graveyard.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1c6a0d19-c7dd-4353-95ad-dfdef2d5d2d9",
		Name:         "Laboratory Drudge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBeginEndStep, func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Kind == game.EventBeginEndStep
			}, "Laboratory Drudge — draw a card if you've cast a spell from a graveyard or activated an ability of a card in a graveyard this turn",
				func(g *game.Game, item *game.StackItem) error {
					for _, ev := range g.EventsThisTurn() {
						if ev.Actor != item.Controller {
							continue
						}
						if (ev.Kind == game.EventCast && ev.OldZone == game.ZoneGraveyard) ||
							(ev.Kind == game.EventActivateAbility && ev.ActivatedFrom == game.ZoneGraveyard) {
							return effects.DrawCards{N: 1}.Apply(effects.NewContext(g, item))
						}
					}
					return nil
				}),
		},
	})
}
