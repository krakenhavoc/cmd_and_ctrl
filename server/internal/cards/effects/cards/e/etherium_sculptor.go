package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Etherium Sculptor — Artifact Creature — Vedalken Artificer {1}{U},
// 1/2:
//
//	"Artifact spells you cast cost {1} less to cast."
//
// The creature-shaped cousin of Foundry Inspector — same predicate
// pair (YourSpell, ArtifactSpell) from cost_modifier.go.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "96b87445-6362-4a17-91b2-cbf203fd03fd",
		Name:         "Etherium Sculptor",
		Completeness: effects.CompletenessFull,
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Artifact spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ArtifactSpell()),
		},
	})
}
