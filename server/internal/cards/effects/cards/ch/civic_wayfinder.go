package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "b9e53414-c21f-4479-b8c1-8b36c3e6e006",
		Name:         "Civic Wayfinder",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Civic Wayfinder — search for a basic land card", effects.MayFetchABasicLandToHand("Civic Wayfinder")),
		},
	})
}
