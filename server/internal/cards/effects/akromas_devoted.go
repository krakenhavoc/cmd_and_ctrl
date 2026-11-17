package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akroma's Devoted — Creature — Human Cleric {3}{W}, 2/4 (Legions, #5609):
//
//	"Cleric creatures have vigilance."
//
// A tribal keyword grant with no "you control" and no "other": every
// Cleric creature at the table has vigilance, the Devoted itself and an
// opponent's Clerics included (CR 702.20b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d264903b-d23e-48c8-95f1-cb1606a705a2",
		Name:         "Akroma's Devoted",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Cleric"}}, "vigilance"),
		},
	})
}
