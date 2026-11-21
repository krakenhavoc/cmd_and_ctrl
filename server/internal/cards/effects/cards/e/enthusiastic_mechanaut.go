package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Enthusiastic Mechanaut — Artifact Creature — Goblin Artificer {U}{R},
// 2/2 (Foundations Commander):
//
//	"Flying
//	 Artifact spells you cast cost {1} less to cast."
//
// Etherium Sculptor's cost reduction on a flyer.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c9722086-23e8-43dc-8a8c-541ff6962318",
		Name:            "Enthusiastic Mechanaut",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			effects.CostsLess(1, "Artifact spells you cast cost {1} less to cast.", effects.YourSpell(), effects.ArtifactSpell()),
		},
	})
}
