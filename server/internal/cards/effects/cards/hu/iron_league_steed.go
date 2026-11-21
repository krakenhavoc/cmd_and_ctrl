package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Iron League Steed — Artifact Creature — Construct {4}, 2/2 (Kaladesh, #5539):
//
//	"Haste
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
		OracleID:        "bafeab74-918b-4fcd-8057-55922dbc1784",
		Name:            "Iron League Steed",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered:       []game.TriggeredAbility{effects.Fabricate("Iron League Steed", 1)},
	})
}
