package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mwonvuli Beast Tracker — Creature — Human Scout {1}{G}{G}, 2/1 (Magic 2013):
//
//	"When this creature enters, search your library for a creature card with
//	 deathtouch, hexproof, reach, or trample and reveal it. Shuffle and put that
//	 card on top."
//
// Moon-Blessed Cleric's tutor-to-top, mandatory this time ("search" is not "you
// may"): the card is found by its printed keywords as a library card has them
// (game.HasKeyword), revealed, then the library is shuffled and the card goes
// on top. A library with no such creature is still searched and shuffled
// (CR 701.23).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b6265654-84c3-469b-9afa-f752fd64abe0",
		Name:         "Mwonvuli Beast Tracker",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Mwonvuli Beast Tracker — search for a creature card with deathtouch, hexproof, reach, or trample", func(g *game.Game, item *game.StackItem) error {
				return effects.SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return c.IsCreature() && (game.HasKeyword(&c, "deathtouch") || game.HasKeyword(&c, "hexproof") ||
							game.HasKeyword(&c, "reach") || game.HasKeyword(&c, "trample"))
					},
					Dest:    game.ZoneLibrary,
					ToTop:   true,
					Limit:   1,
					Reveal:  true,
					Shuffle: true,
					Reason:  "Mwonvuli Beast Tracker — a creature card with deathtouch, hexproof, reach, or trample",
				}.Apply(effects.NewContext(g, item))
			}),
		},
	})
}
