package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// The Five Doctors — Sorcery {5}{G} (Doctor Who):
//
//	"Kicker {5}
//	 Search your library and/or graveyard for up to five Doctor cards, reveal
//	 them, and put them into your hand. If you search your library this way,
//	 shuffle. If this spell was kicked, put those cards onto the battlefield
//	 instead of putting them into your hand."
//
// One search over the library and the graveyard (SearchLibrary.AlsoGraveyard,
// Finale of Devastation's), up to five cards with the Doctor subtype, revealed.
// Kicked, the destination is the battlefield instead of the hand. The shuffle
// is the library search's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:      "dbdd93d2-f2c6-43f9-93e6-2185c7df83bc",
		Name:          "The Five Doctors",
		Completeness:  CompletenessFull,
		Purpose:       game.Purpose{Tutors: 5},
		OptionalCosts: []game.AdditionalCost{Kicker("{5}")},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			dest := game.ZoneHand
			if ctx.WasKicked() {
				dest = game.ZoneBattlefield
			}
			return SearchLibrary{
				Player:        item.Controller,
				Predicate:     func(c game.Card) bool { return c.HasSubtypeWord("Doctor") },
				Dest:          dest,
				Limit:         5,
				UpTo:          true,
				Reveal:        true,
				Shuffle:       true,
				AlsoGraveyard: true,
				Reason:        "The Five Doctors — up to five Doctor cards",
				Source:        item.SourceCardID,
			}.Apply(ctx)
		},
	})
}
