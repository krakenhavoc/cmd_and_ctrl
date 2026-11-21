package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Borderland Ranger — Creature — Human Scout Ranger {2}{G}, 2/2 (Magic 2010, #5585):
//
//	"When this creature enters, you may search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle."
//
// Sylvan Ranger's enters trigger: a "you may" search for a basic land card,
// revealed and put into hand, then the library is shuffled.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ffe25048-0bd4-49b5-aecb-2b1b57fb8d8b",
		Name:         "Borderland Ranger",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tutors: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Borderland Ranger — search for a basic land card", effects.MayFetchABasicLandToHand("Borderland Ranger")),
		},
	})
}
