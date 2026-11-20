package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "1dff7c82-e49f-45e8-8f66-d5801f420fa6",
		Name:            "Nim Shrieker",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Static:          []game.StaticAbility{mrdB5PowerPerArtifact()},
	})
}
