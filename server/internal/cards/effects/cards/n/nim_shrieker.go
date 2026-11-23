package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nim Shrieker — Creature — Zombie {3}{B}, 0/1 (Mirrodin, #5606):
//
//	"Flying
//	 This creature gets +1/+0 for each artifact you control."
//
// Printed flying, and Nim Lasher's count (mrdB5PowerPerArtifact): a
// layer 7c boost read through the layers on every recompute.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1dff7c82-e49f-45e8-8f66-d5801f420fa6",
		Name:            "Nim Shrieker",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{effects.MrdB5PowerPerArtifact()},
	})
}
