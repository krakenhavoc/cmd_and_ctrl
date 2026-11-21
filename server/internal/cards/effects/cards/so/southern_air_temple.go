package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Southern Air Temple — Legendary Enchantment — Shrine {3}{W} (Avatar:
// The Last Airbender, #4459):
//
//	"When Southern Air Temple enters, put X +1/+1 counters on each
//	 creature you control, where X is the number of Shrines you control.
//	 Whenever another Shrine you control enters, put a +1/+1 counter on
//	 each creature you control."
//
// Two triggers, two rows. X is counted as the entry trigger resolves
// (CR 608.2h), the Temple itself among the Shrines while it is still
// yours; each creature then gets its X counters as ONE placement
// (b11PutCountersOnEachCreatureYouControl), so a counter replacement
// applies to each creature once. The set of creatures is read at
// resolution, so a creature that entered in response gets counters too.
// The second row is Sanctum of Fruitful Harvest's Shrine read on an
// entering permanent: another Shrine that entered under your control.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "987c1144-69f6-4441-b0a4-d23f931df81f",
		Name:         "Southern Air Temple",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Southern Air Temple — put X +1/+1 counters on each creature you control, where X is the number of Shrines you control",
				func(g *game.Game, item *game.StackItem) error {
					shrines := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype("Shrine") })
					if shrines == 0 {
						return nil
					}
					return effects.B11PutCountersOnEachCreatureYouControl(g, item, shrines)
				}),
			effects.On(game.EventETB, southernAirTempleAnotherShrineEntered,
				"Southern Air Temple — put a +1/+1 counter on each creature you control",
				effects.B11PutCounterOnEachCreatureYouControl),
		},
	})
}

// southernAirTempleAnotherShrineEntered is "another Shrine you control
// enters".
func southernAirTempleAnotherShrineEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Shrine"))
}
