package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tangle Golem — Artifact Creature — Golem {7} (Darksteel, #5605):
//
//	"Affinity for Forests (This spell costs {1} less to cast for each
//	 Forest you control.)"
//
// Spire Golem's shape: affinity is the shared self cost modifier counting
// the Forests its caster controls (CR 702.41a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "733c6e30-3c0c-41f0-b1e9-4c2eb76f7b82",
		Name:         "Tangle Golem",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for Forests", And(Land(), Subtype("Forest"))),
		},
	})
}
