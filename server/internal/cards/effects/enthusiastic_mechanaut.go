package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "c9722086-23e8-43dc-8a8c-541ff6962318",
		Name:            "Enthusiastic Mechanaut",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		CostModifiers: []game.CostModifier{
			CostsLess(1, "Artifact spells you cast cost {1} less to cast.", YourSpell(), ArtifactSpell()),
		},
	})
}
