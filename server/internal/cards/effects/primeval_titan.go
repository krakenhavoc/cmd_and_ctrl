package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Primeval Titan — Creature — Giant {4}{G}{G}, 6/6:
//
//	"Trample
//	 Whenever this creature enters or attacks, you may search your
//	 library for up to two land cards, put them onto the battlefield
//	 tapped, then shuffle."
//
// One printed ability with two trigger conditions, so one
// TriggeredAbility watching both kinds (Sun Titan's shape), and a
// SearchLibrary to the battlefield with TappedOnEntry. "Up to two" is
// the chooser's zero minimum and "you may" is Optional, so declining
// and finding nothing are both answers. The lands are any land cards,
// not basics, and they are shuffled in afterwards.
//
// The Purpose counts the enters half (two lands), which is the one the
// bot can rely on; the attack half is the same search and is not a
// second declaration.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "ae83ef2c-960f-4c5b-97cc-52465c687c18",
		Name:            "Primeval Titan",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Purpose:         game.Purpose{Lands: 2},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersOrAttacks("Primeval Titan — search for up to two land cards",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:        item.Controller,
						Predicate:     isLandCard,
						Dest:          game.ZoneBattlefield,
						Limit:         2,
						Reveal:        true,
						Shuffle:       true,
						Optional:      true,
						TappedOnEntry: true,
						Source:        item.SourceCardID,
						Reason:        "Primeval Titan — up to two land cards, onto the battlefield tapped",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
