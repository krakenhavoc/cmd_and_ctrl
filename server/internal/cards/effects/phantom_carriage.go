package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Phantom Carriage — Creature — Spirit {4}{U}{U}, 4/4 (Innistrad: Midnight
// Hunt, S82):
//
//	"Flying
//	 When this creature enters, you may search your library for a card
//	 with flashback or disturb, put it into your graveyard, then shuffle."
//
// A "may search" with a stated quality: the searcher may decline, and may
// find nothing (CR 701.23b). The quality is the card's printed keyword
// (hasFlashbackOrDisturb): flashback read from the printed text, disturb
// from the keyword, so a card a catalog entry does not run counts too.
// The card goes straight to the graveyard, and the library is shuffled
// either way.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d95e7282-b298-4410-baac-01766d6852e3",
		Name:            "Phantom Carriage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Purpose:         game.Purpose{SelfMillTutor: 1},
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Phantom Carriage — you may search your library for a card with flashback or disturb, put it into your graveyard, then shuffle",
				func(g *game.Game, item *game.StackItem) error {
					return SearchLibrary{
						Player:    item.Controller,
						Predicate: hasFlashbackOrDisturb,
						Dest:      game.ZoneGraveyard,
						Limit:     1,
						Shuffle:   true,
						Optional:  true,
						Source:    item.SourceCardID,
						Reason:    "Phantom Carriage — put a card with flashback or disturb into your graveyard",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
