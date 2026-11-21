package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wistful Selkie — Creature — Merfolk Wizard {G/U}{G/U}{G/U}, 2/2 (Eventide, #5590):
//
//	"When this creature enters, draw a card."
//
// Wall of Blossoms' enters trigger.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4503d0fb-842d-49f1-9769-8509ff0f2ce1",
		Name:         "Wistful Selkie",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Wistful Selkie — draw a card", effects.Do(effects.DrawCards{N: 1})),
		},
	})
}
