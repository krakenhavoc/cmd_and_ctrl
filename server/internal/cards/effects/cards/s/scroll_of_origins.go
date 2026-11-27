package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Scroll of Origins — Artifact {2} (Saviors of Kamigawa, #5601):
//
//	"{2}, {T}: Draw a card if you have seven or more cards in hand."
//
// The seven cards are an "if" on the effect, not a restriction on activating:
// the ability can be activated with fewer, and the hand is counted as it
// resolves, so it draws nothing if the hand has shrunk below seven.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6974c433-a522-40bb-8791-375fcc242234",
		Name:         "Scroll of Origins",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}, {T}: Draw a card if you have seven or more cards in hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost()),
			Effect:  effects.OldB585DrawIfSevenCards,
		}},
	})
}
