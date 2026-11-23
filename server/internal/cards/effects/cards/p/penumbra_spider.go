package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Penumbra Spider — Creature — Spider {2}{G}{G}, 2/4 (Time Spiral, #5596):
//
//	"Reach
//	 When this creature dies, create a 2/4 black Spider creature token
//	 with reach."
//
// The dies trigger creates the black 2/4 Spider under the controller of
// the Spider that died, read from its last-known information (CR 603.10a).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "47678f44-b595-492e-af00-4e73f941e24e",
		Name:            "Penumbra Spider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Penumbra Spider — create a 2/4 black Spider creature token with reach",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/4 black Spider with reach"), N: 1})),
		},
	})
}
