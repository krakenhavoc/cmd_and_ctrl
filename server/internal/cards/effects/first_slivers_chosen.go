package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// First Sliver's Chosen — Creature — Sliver {4}{W}, 3/3 (Modern Horizons, #5519):
//
//	"Sliver creatures you control have exalted. (Whenever a creature you
//	 control attacks alone, it gets +1/+1 until end of turn for each
//	 instance of exalted among permanents you control.)"
//
// A layer-6 grant of the exalted keyword. Exalted is cumulative, so each
// Sliver you control is one more instance and a lone attacker gets +1/+1
// for each (CR 702.83b); two Chosen give each Sliver two instances.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f664596a-9ad1-4db3-95b0-457698ad2f66",
		Name:         "First Sliver's Chosen",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}, YoursOnly: true}, game.KeywordExalted),
		},
	})
}
