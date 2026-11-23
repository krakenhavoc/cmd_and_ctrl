package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Court Homunculus — Artifact Creature — Homunculus {W}, 1/1 (Conflux,
// #5587):
//
//	"This creature gets +1/+1 as long as you control another artifact."
//
// Magitek Infantry's condition (another artifact you control, the
// Homunculus itself not counted) at +1/+1, applied in layer 7c (CR
// 613.4c) and read continuously.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b597e801-2648-48b9-8a21-02baa13e37af",
		Name:         "Court Homunculus",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "This creature gets +1/+1 as long as you control another artifact.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: effects.MagitekInfantryWhileAnotherArtifact,
			Apply:     effects.PumpBy(1, 1),
		}},
	})
}
