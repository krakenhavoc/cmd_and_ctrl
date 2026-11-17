package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f2ffc999-609f-4424-8d40-ebc042357a40",
		Name:         "Powerstone Engineer",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Powerstone Engineer — create a tapped Powerstone token",
				Do(CreateTokenAdvanced{Spec: Token(PowerstoneToken()).EntersTapped(), N: 1})),
		},
	})
}
