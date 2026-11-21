package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inventor's Apprentice — Creature — Human Artificer {R}, 1/2 (Kaladesh, #5539):
//
//	"This creature gets +1/+1 as long as you control an artifact."
//
// Goblin Tomb Raider's shape: one layer 7c +1/+1 (CR 613.4c) on this
// creature alone, gated on Cloudsculpt Technician's "you control an
// artifact" check, read after the layers so it ends the moment the last
// artifact leaves.
//
// No simplification.
func init() {
	whileArtifact := effects.WhileYouHave(effects.CloudsculptControlsAnArtifact, effects.SelfCreature)
	effects.Register(effects.Spec{
		OracleID:     "ed8d54fa-ec02-45d6-b70e-4c7e1c66efd6",
		Name:         "Inventor's Apprentice",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+1 as long as you control an artifact.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileArtifact,
			Apply:     effects.PumpBy(1, 1),
		}},
	})
}
