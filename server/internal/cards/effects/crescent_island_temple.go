package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "2886f95e-c4e8-4770-b7de-40985f336cfb",
		Name:         "Crescent Island Temple",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Crescent Island Temple — for each Shrine you control, create a 1/1 red Monk creature token with prowess",
				func(g *game.Game, item *game.StackItem) error {
					shrines := countControlled(g, item.Controller, func(c game.Card) bool { return c.HasSubtype("Shrine") })
					if shrines == 0 {
						return nil
					}
					return CreateToken{Template: TokenCard("1/1 red Monk with prowess"), N: shrines}.Apply(NewContext(g, item))
				}),
			On(game.EventETB, crescentIslandTempleAnotherShrineEntered,
				"Crescent Island Temple — create a 1/1 red Monk creature token with prowess",
				Do(CreateToken{Template: TokenCard("1/1 red Monk with prowess"), N: 1})),
		},
	})
}

// crescentIslandTempleAnotherShrineEntered is "another Shrine you
// control enters".
func crescentIslandTempleAnotherShrineEntered(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
	c, ok := enteredUnderYourControl(ev, source, g, true)
	return ok && c.HasSubtype(source.TextSubtype("Shrine"))
}
