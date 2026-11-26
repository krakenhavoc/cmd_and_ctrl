package cardsa

import (
	"strconv"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Angelheart Vial — Artifact {5} (Rise of the Eldrazi, S82 OLD batch 482,
// #5581):
//
//	"Whenever you're dealt damage, you may put that many charge counters on
//	 this artifact.
//	 {2}, {T}, Remove four charge counters from this artifact: You gain 2 life
//	 and draw a card."
//
// Sun Droplet's damage trigger made optional (CR 603.5): the amount is the
// damage dealt, read off the event. The counters are a cost, paid as the
// ability is activated.
//
// No simplification.
func init() {
	dealt := game.TriggeredAbility{
		Watches: []game.EventKind{game.EventDealDamage},
		AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
			_, ok := effects.B43YouWereDealtDamage(ev, source, g)
			return ok
		},
		Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
			n, _ := effects.B43YouWereDealtDamage(ev, source, g)
			return game.NewTriggeredItem(source, "Angelheart Vial — put "+strconv.Itoa(n)+" charge counters on it")
		},
		Effect: func(g *game.Game, item *game.StackItem) error {
			if item.Trigger == nil {
				return nil
			}
			return effects.CounterOnThis{Kind: game.CounterCharge, N: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
		},
	}
	effects.Register(effects.Spec{
		OracleID:     "efbc9b9c-2b8c-4121-b232-350c80f12a2a",
		Name:         "Angelheart Vial",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(dealt, "Angelheart Vial — put the counters on it?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}, Remove four charge counters from this artifact: You gain 2 life and draw a card.",
			Purpose: game.Purpose{Draws: 1, LifeGain: 2, Answers: game.AnswerValue},
			Cost: game.AbilityCost{
				Mana:           "{2}",
				Tap:            true,
				RemoveCounters: &game.CounterRemovalCost{Counter: game.CounterCharge, N: 4},
			},
			Effect: effects.Do(effects.GainLife{Amount: 2}, effects.DrawCards{N: 1}),
		}},
	})
}
