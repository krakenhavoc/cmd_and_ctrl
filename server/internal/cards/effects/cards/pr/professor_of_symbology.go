package cardspr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Professor of Symbology — Creature — Kor Cleric {1}{W}, 2/1 (EDHREC
// rank 26247):
//
//	"When this creature enters, learn. (You may reveal a Lesson card you
//	 own from outside the game and put it into your hand, or discard a
//	 card to draw a card.)"
//
// An enters trigger whose instruction is learn (CR 701.48a, learn.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "1a8ae444-6237-4591-97ab-f43e181b2f4f",
		Name:         "Professor of Symbology",
		Purpose:      game.Purpose{Learn: 1},
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Professor of Symbology — learn", effects.Do(effects.Learn{})),
		},
	})
}
