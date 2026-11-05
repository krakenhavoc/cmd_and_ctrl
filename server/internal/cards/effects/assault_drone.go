package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Assault Drone — Artifact Creature — Borg {3}{B}, 2/4 (Star Trek):
//
//	Artifact creatures you control get +1/+0.
//
// A layer 7c anthem over the artifact creatures its controller controls,
// itself included (no "other"), read off each permanent's effective types.
//
// No simplification.
func init() {
	anthem := b16Anthem(func(target *game.Card, _ *game.Game, source *game.Card) bool {
		return target.Controller == source.Controller && target.IsArtifact() && target.IsCreature()
	}, 1, 0)
	anthem.Label = "Artifact creatures you control get +1/+0."
	Register(Spec{
		OracleID:     "18832d7a-9a1f-45f1-a0f9-8e768a36af07",
		Name:         "Assault Drone",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{anthem},
	})
}
