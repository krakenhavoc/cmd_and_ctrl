package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Powerstone Engineer — Creature — Human Artificer {1}{W}, 2/1 (The
// Brothers' War, #5482):
//
//	"When this creature dies, create a tapped Powerstone token."
//
// The dies trigger is the shared WhenThisDies; the token is yours (the
// controller of the ability) and enters tapped.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f2ffc999-609f-4424-8d40-ebc042357a40",
		Name:         "Powerstone Engineer",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Powerstone Engineer — create a tapped Powerstone token",
				effects.Do(effects.CreateTokenAdvanced{Spec: effects.Token(effects.PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
