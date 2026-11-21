package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Koilos Roc — Creature — Bird {4}{U}, 3/3 (The Brothers' War, #5482):
//
//	"Flash
//	 Flying
//	 When this creature enters, create a tapped Powerstone token."
//
// Flash and flying are keywords; the enters trigger is Stone Retrieval
// Unit's.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "0956c0f5-ff1d-43d5-88de-4c2d603cf798",
		Name:            "Koilos Roc",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Koilos Roc — create a tapped Powerstone token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
