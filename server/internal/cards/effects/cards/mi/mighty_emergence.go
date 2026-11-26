package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mighty Emergence — {2}{G} Enchantment (Shards of Alara, S82 OLD b514):
//
//	"Whenever a creature you control with power 5 or greater enters, you may
//	 put two +1/+1 counters on it."
//
// Power is read off the creature as it entered (CR 603.2, layered view), so a
// creature that enters with counters counts them. The counters go on the
// creature that entered, if it is still on the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e7f038af-977a-42ff-bebb-718b7490188a",
		Name:         "Mighty Emergence",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventETB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				c, ok := effects.EnteredUnderYourControl(ev, source, g, false)
				return ok && c.IsCreature() && c.CurrentPower() >= 5
			}, "Mighty Emergence — put two +1/+1 counters on that creature", func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				info, ok := ctx.TriggeringPermanent()
				if !ok || info.Left || item.Trigger == nil {
					return nil
				}
				return effects.AddCounter{Target: item.Trigger.Event.CardID, Kind: game.CounterPlusOne, N: 2}.Apply(ctx)
			}), "Mighty Emergence — put two +1/+1 counters on it?"),
		},
	})
}
