package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Northern Air Temple — Legendary Enchantment — Shrine {B} (Avatar: The
// Last Airbender, #4459):
//
//	"When Northern Air Temple enters, each opponent loses X life and you
//	 gain X life, where X is the number of Shrines you control.
//	 Whenever another Shrine you control enters, each opponent loses 1
//	 life and you gain 1 life."
//
// Two triggers, two rows. X is counted as the entry trigger resolves
// (CR 608.2h), the Temple itself among the Shrines while it is still
// yours, and you gain X however many opponents lost it (b40's drain).
// The second row is a footprinted drain on another Shrine entering under
// your control; "Shrine" is a word of the Temple's text (CR 612.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "28a2ad5b-d637-4634-8d4b-889e1fc88319",
		Name:         "Northern Air Temple",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Northern Air Temple — each opponent loses X life and you gain X life, where X is the number of Shrines you control",
				func(g *game.Game, item *game.StackItem) error {
					shrines := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype("Shrine") })
					if shrines == 0 {
						return nil
					}
					return b40DrainEachOpponent(shrines)(g, item)
				}),
			On(game.EventETB, northernAirTempleAnotherShrineEntered,
				"Northern Air Temple — each opponent loses 1 life and you gain 1 life",
				Do(EachOpponentLosesLife{N: 1}, GainLife{Amount: 1})),
		},
	})
}

// northernAirTempleAnotherShrineEntered is "another Shrine you control
// enters".
func northernAirTempleAnotherShrineEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Shrine"))
}
