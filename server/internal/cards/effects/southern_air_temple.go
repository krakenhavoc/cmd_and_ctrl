package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "987c1144-69f6-4441-b0a4-d23f931df81f",
		Name:         "Southern Air Temple",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Southern Air Temple — put X +1/+1 counters on each creature you control, where X is the number of Shrines you control",
				func(g *game.Game, item *game.StackItem) error {
					shrines := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype("Shrine") })
					if shrines == 0 {
						return nil
					}
					return b11PutCountersOnEachCreatureYouControl(g, item, shrines)
				}),
			On(game.EventETB, southernAirTempleAnotherShrineEntered,
				"Southern Air Temple — put a +1/+1 counter on each creature you control",
				b11PutCounterOnEachCreatureYouControl),
		},
	})
}

// southernAirTempleAnotherShrineEntered is "another Shrine you control
// enters".
func southernAirTempleAnotherShrineEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Shrine"))
}
