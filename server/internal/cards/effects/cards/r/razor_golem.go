package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Razor Golem — Artifact Creature — Golem {6} (Darksteel, #5605):
//
//	"Affinity for Plains (This spell costs {1} less to cast for each
//	 Plains you control.)
//	 Vigilance"
//
// Spire Golem's shape: affinity is the shared self cost modifier counting
// the Plains its caster controls (CR 702.41a), and vigilance is a printed
// keyword.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "c7b2188a-7571-48b8-ac66-ca741bf2e779",
		Name:            "Razor Golem",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		SelfCostModifiers: []game.CostModifier{
			effects.AffinityFor("Affinity for Plains", effects.And(effects.Land(), effects.Subtype("Plains"))),
		},
	})
}
