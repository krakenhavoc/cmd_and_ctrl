package cardsh

import (
	"github.com/google/uuid"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Haruspex — Creature — Tyranid {3}{G}, 2/2:
//
//	"Rapacious Hunger — Whenever another creature dies, put a +1/+1 counter
//	 on this creature.
//	 Devouring Monster — {T}, Remove X +1/+1 counters from this creature: Add
//	 X mana of any one color."
//
// The second ability is a mana ability (no target, adds mana, CR 605.1a), with
// the variable removal as part of its cost. X may be 0.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2309436e-3442-46d7-8ffc-9dc4a2314ae9",
		Name:         "Haruspex",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, lki game.Characteristic, g *game.Game) bool {
				return effects.AnotherCreatureDied(ev, source, lki, g)
			}, "Haruspex — put a +1/+1 counter on this creature", effects.Do(effects.CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
		},
		ManaAbilities: []effects.ManaAbility{{
			Cost: effects.ManaAbilityCost{
				Tap:            true,
				RemoveCounters: effects.RemoveCountersXFromThis(game.CounterPlusOne, 0).RemoveCounters,
			},
			ProducedForPaid: func(_ *game.Game, _, _ uuid.UUID, paid game.PaidCost) game.ManaOutput {
				return game.ManaOutput{Amount: game.ManaAmount{Options: game.AllColors, N: paid.CountersRemoved, OneColor: true}}
			},
			Label: "{T}, Remove X +1/+1 counters from this creature: Add X mana of any one color",
		}},
	})
}
