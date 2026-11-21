package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Venerated Loxodon — Creature — Elephant Cleric {4}{W} 4/4:
//
//	"Convoke
//	 When this creature enters, put a +1/+1 counter on each creature that
//	 convoked it."
//
// Zephyr Singer's shape with +1/+1 counters: the creatures that convoked it
// ride the permanent (CR 400.7d, CastProvenance.Convoked) and are read as
// the trigger is built, so the counters still land if the Loxodon has left.
// A creature that has left the battlefield since is skipped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d157a9df-59bc-40ad-b62b-43f973f14be1",
		Name:         "Venerated Loxodon",
		Completeness: effects.CompletenessFull,
		TapCost:      effects.Convoke(),
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventETB},
			AppliesTo: effects.Self,
			Key:       veneratedLoxodonLabel,
			Build: func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
				item := game.NewTriggeredItem(source, veneratedLoxodonLabel)
				item.Payload = effects.ConvokersPayload(source)
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				for _, id := range ctx.PayloadCards() {
					if !effects.OnBattlefield(g, id) {
						continue
					}
					if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
						return err
					}
				}
				return nil
			},
		}},
	})
}

const veneratedLoxodonLabel = "Venerated Loxodon — put a +1/+1 counter on each creature that convoked it"
