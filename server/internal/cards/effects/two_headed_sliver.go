package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Two-Headed Sliver — Creature — Sliver {1}{R}, 1/1 (Time Spiral, #5596):
//
//	"All Sliver creatures have menace. (They can't be blocked except by
//	 two or more creatures.)"
//
// A layer-6 keyword grant (CR 613.1f) to every Sliver creature at the
// table, an opponent's included. Menace is enforced at the block
// declaration (CR 702.111b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "be857d6e-5bf2-491b-ab46-963b4d45adee",
		Name:         "Two-Headed Sliver",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Sliver"}}, "menace"),
		},
	})
}
