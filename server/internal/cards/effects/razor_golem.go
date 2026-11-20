package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "c7b2188a-7571-48b8-ac66-ca741bf2e779",
		Name:            "Razor Golem",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Plains", And(Land(), Subtype("Plains"))),
		},
	})
}
