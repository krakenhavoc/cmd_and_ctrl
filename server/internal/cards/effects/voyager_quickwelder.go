package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Voyager Quickwelder — Artifact Creature — Robot Artificer {2}{W}, 2/4
// (Aetherdrift, #4464):
//
//	"Artifact spells you cast cost {1} less to cast."
//
// Foundry Inspector's cost modifier (CR 601.2f): only its controller's
// artifact spells, and only generic mana comes off.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87e8c044-5f83-437d-a89e-8a95559045f6",
		Name:         "Voyager Quickwelder",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Artifact spells you cast cost {1} less to cast.", YourSpell(), ArtifactSpell()),
		},
	})
}
