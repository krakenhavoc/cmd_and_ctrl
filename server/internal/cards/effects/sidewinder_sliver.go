package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sidewinder Sliver — Creature — Sliver {W}, 1/1 (Time Spiral, #5596):
//
//	"All Sliver creatures have flanking. (Whenever a creature without
//	 flanking blocks a Sliver, the blocking creature gets -1/-1 until
//	 end of turn.)"
//
// A layer-6 grant (CR 613.1f) of flanking to every Sliver creature at the
// table. Flanking is the engine's triggered keyword (CR 702.25a): a
// granted instance triggers like a printed one, and a blocker that has
// flanking itself is not shrunk.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "de8ccd32-f537-4a11-8dc4-e48836a0b139",
		Name:         "Sidewinder Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, game.KeywordFlanking),
		},
	})
}
