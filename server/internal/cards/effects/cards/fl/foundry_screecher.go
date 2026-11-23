package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Foundry Screecher — Creature — Bat {2}{B}, 2/1 (Kaladesh, #5539):
//
//	"Flying
//	 This creature gets +1/+0 as long as you control an artifact."
//
// Flying is printed. The pump is Goblin Tomb Raider's: a layer 7c +1/+0
// (CR 613.4c) on this creature alone while its controller controls an
// artifact.
//
// No simplification.
func init() {
	whileArtifact := effects.WhileYouHave(effects.CloudsculptControlsAnArtifact, effects.SelfCreature)
	effects.Register(effects.Spec{
		OracleID:        "0e0bef57-d977-48d1-be78-9a4f0aacd9c0",
		Name:            "Foundry Screecher",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static: []game.StaticAbility{{
			Label:     "Gets +1/+0 as long as you control an artifact.",
			Layer:     game.Layer7PT,
			SubLayer:  game.SubLayer7C_Modify,
			AppliesTo: whileArtifact,
			Apply:     effects.PumpBy(1, 0),
		}},
	})
}
