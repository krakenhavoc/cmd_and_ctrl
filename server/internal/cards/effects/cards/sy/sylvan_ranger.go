package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sylvan Ranger — Creature — Elf Scout Ranger {1}{G}, 1/1:
//
//	"When this creature enters, you may search your library for a basic
//	 land card, reveal it, put it into your hand, then shuffle."
//
// The "you may" is the search prompt's decline, which skips the shuffle
// too; the pick is revealed and goes to hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "29a90fda-541e-4923-b3ed-e2b35d02c74b",
		Name:         "Sylvan Ranger",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Sylvan Ranger — search for a basic land card", effects.MayFetchABasicLandToHand("Sylvan Ranger")),
		},
	})
}
