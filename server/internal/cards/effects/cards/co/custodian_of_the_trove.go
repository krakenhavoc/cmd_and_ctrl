package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Custodian of the Trove — Artifact Creature — Golem {3}, 2/5 (Dragons of Tarkir, #5551):
//
//	"Defender
//	 This creature enters tapped."
//
// Defender is a printed keyword and the tapped entry a replacement on itself
// (CR 614.12).
//
// No simplification.

func init() {
	effects.Register(effects.Spec{
		OracleID:        "0d060d19-9c8b-4831-a9be-aa2c98b6858a",
		Name:            "Custodian of the Trove",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"defender"},
		Replacements:    []game.ReplacementEffect{effects.SelfEntersTapped()},
	})
}
