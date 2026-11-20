package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nim Lasher — Creature — Zombie {2}{B}, 1/1 (Mirrodin, #5606):
//
//	"This creature gets +1/+0 for each artifact you control."
//
// Guidelight Synergist's count (mrdB5PowerPerArtifact): a layer 7c
// boost read through the layers on every recompute.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0203e250-d60e-46c6-9c58-d620f0377be0",
		Name:         "Nim Lasher",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{mrdB5PowerPerArtifact()},
	})
}
