package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Fen Hauler — Creature — Insect {6}{B}, 5/5:
//
//	"Improvise ... This creature can't be blocked by artifact creatures."
//
// Improvise is Spec.TapCost; the block restriction is a CantBeBlockedBy
// rule on itself (Questing Beast's shape). No simplification.
func init() {
	Register(Spec{
		OracleID:     "0f909cfd-c643-48d2-983a-da2e38cd2bff",
		Name:         "Fen Hauler",
		Completeness: CompletenessFull,
		TapCost:      Improvise(),
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), And(Creature(), Artifact()), "artifact creatures"),
		},
	})
}
