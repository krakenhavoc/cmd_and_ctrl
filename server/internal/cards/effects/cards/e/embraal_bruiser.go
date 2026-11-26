package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Embraal Bruiser — Creature — Human Warrior {1}{B}, 3/1 (Kaladesh, #5539):
//
//	"This creature enters tapped.
//	 This creature has menace as long as you control an artifact."
//
// Enters tapped is the usual self-tapping replacement; menace is a layer 6 grant on this
// creature alone while its controller controls an artifact, read after the
// layers (Foundry Screecher's gate). It ends the moment the last artifact
// leaves.
//
// No simplification.
func init() {
	menace := effects.KeywordGrant(effects.WhileYouHave(effects.CloudsculptControlsAnArtifact, effects.SelfCreature), "menace")
	menace.Label = "Has menace as long as you control an artifact."
	effects.Register(effects.Spec{
		OracleID:     "f432998f-7737-4f41-86b5-a4372bc5db2b",
		Name:         "Embraal Bruiser",
		Completeness: effects.CompletenessFull,
		Replacements: []game.ReplacementEffect{effects.SelfEntersTapped()},
		Static:       []game.StaticAbility{menace},
	})
}
