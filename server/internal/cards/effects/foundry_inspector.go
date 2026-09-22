package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foundry Inspector — Artifact Creature — Construct {3}, 3/2
// (EDHREC rank 252):
//
//	"Artifact spells you cast cost {1} less to cast."
//
// The affinity-adjacent cost reducer — CostsLess with YourSpell() and
// the new ArtifactSpell() predicate (cost_modifier.go), the same
// vocabulary the roadmap's cost-modification group (#93) unblocked
// for the whole batch.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "18f22960-87ec-43cd-82ea-ec5cabf49ad3",
		Name:         "Foundry Inspector",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Artifact spells you cast cost {1} less to cast.", YourSpell(), ArtifactSpell()),
		},
	})
}
