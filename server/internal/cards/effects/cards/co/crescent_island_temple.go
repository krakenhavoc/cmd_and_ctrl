package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Crescent Island Temple — Legendary Enchantment — Shrine {3}{R}
// (Avatar: The Last Airbender, #4459):
//
//	"When Crescent Island Temple enters, for each Shrine you control,
//	 create a 1/1 red Monk creature token with prowess.
//	 Whenever another Shrine you control enters, create a 1/1 red Monk
//	 creature token with prowess."
//
// Two triggers, two rows. The Shrines are counted as the entry trigger
// resolves (CR 608.2h), the Temple itself among them while it is still
// yours. The second row is a footprinted token on another Shrine
// entering under your control; "Shrine" is a word of the Temple's text
// (CR 612.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "2886f95e-c4e8-4770-b7de-40985f336cfb",
		Name:         "Crescent Island Temple",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Crescent Island Temple — for each Shrine you control, create a 1/1 red Monk creature token with prowess",
				func(g *game.Game, item *game.StackItem) error {
					shrines := effects.CountControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype("Shrine") })
					if shrines == 0 {
						return nil
					}
					return effects.CreateToken{Template: effects.TokenCard("1/1 red Monk with prowess"), N: shrines}.Apply(effects.NewContext(g, item))
				}),
			effects.On(game.EventETB, crescentIslandTempleAnotherShrineEntered,
				"Crescent Island Temple — create a 1/1 red Monk creature token with prowess",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("1/1 red Monk with prowess"), N: 1})),
		},
	})
}

// crescentIslandTempleAnotherShrineEntered is "another Shrine you
// control enters".
func crescentIslandTempleAnotherShrineEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := effects.EnteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Shrine"))
}
