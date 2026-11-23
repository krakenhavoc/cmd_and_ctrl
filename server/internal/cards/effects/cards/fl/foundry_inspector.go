package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "18f22960-87ec-43cd-82ea-ec5cabf49ad3",
		Name:         "Foundry Inspector",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Artifact spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ArtifactSpell()),
		},
	})
}
