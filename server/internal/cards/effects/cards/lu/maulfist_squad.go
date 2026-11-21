package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Maulfist Squad — Creature — Human Artificer {3}{B}, 3/1 (Kaladesh, #5539):
//
//	"Menace
//	 Fabricate 1 (When this creature enters, put a +1/+1 counter on it
//	 or create a 1/1 colorless Servo artifact creature token.)"
//
// The printed keyword is read by the engine. Fabricate is the shared keyword constructor (fabricate_keyword.go,
// CR 702.123a): the enters trigger asks counters or Servos as it
// resolves, and makes the Servos without asking once this creature has
// left the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "2a8ec988-b513-4e66-b7f5-b2196b4bbff2",
		Name:            "Maulfist Squad",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"menace"},
		Triggered:       []game.TriggeredAbility{effects.Fabricate("Maulfist Squad", 1)},
	})
}
