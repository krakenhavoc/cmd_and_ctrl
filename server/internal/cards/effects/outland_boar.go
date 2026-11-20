package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Outland Boar — Creature — Boar {2}{R}{G}, 4/4 (Aether Revolt,
// #5536):
//
//	"This creature can't be blocked by creatures with power 2 or less."
//
// Steel Leaf Champion's evasion: the blocker's power is read as blockers
// are declared (CR 509.1b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fdb109f6-d185-4d1e-92b9-768e77562f55",
		Name:         "Outland Boar",
		Completeness: CompletenessFull,
		BlockRules: []game.BlockRule{
			CantBeBlockedBy(OnSelf(), PowerLE(2), "creatures with power 2 or less"),
		},
	})
}
