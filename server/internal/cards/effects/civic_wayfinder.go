package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Civic Wayfinder — Creature — Elf Druid Warrior {2}{G}, 2/2 (Ravnica:
// City of Guilds, #5600):
//
//	"When this creature enters, you may search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle."
//
// Sylvan Ranger's enters trigger (mayFetchABasicLandToHand): the "you may"
// is the search prompt's decline, which skips the shuffle too.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9e53414-c21f-4479-b8c1-8b36c3e6e006",
		Name:         "Civic Wayfinder",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Civic Wayfinder — search for a basic land card", mayFetchABasicLandToHand("Civic Wayfinder")),
		},
	})
}
