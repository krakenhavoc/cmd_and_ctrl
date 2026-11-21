package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Frogmyr Enforcer — Artifact Creature — Frog Myr {7}, 4/4 (Modern Horizons 3):
//
//	"Prototype {3}{R} — 2/2
//	 Affinity for artifacts"
//
// Prototype is the card's second face (CR 718, game/prototype.go); the affinity
// reduces whichever cost the spell is cast for, as CR 702.41a says.
func init() {
	Register(Spec{
		OracleID:     "ccc7de36-aafc-4d87-80f7-fc2effc976eb",
		Name:         "Frogmyr Enforcer",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			AffinityFor("Affinity for artifacts", Artifact()),
		},
	})
}
