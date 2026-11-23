package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Skittering Surveyor — Artifact Creature — Construct {3}, 1/2:
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
		OracleID:     "85364398-923c-4ee9-8519-576fce1c26f7",
		Name:         "Skittering Surveyor",
		Purpose:      game.Purpose{Tutors: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Skittering Surveyor — search for a basic land card", effects.MayFetchABasicLandToHand("Skittering Surveyor")),
		},
	})
}
