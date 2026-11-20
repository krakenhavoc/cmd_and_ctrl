package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunrise Sovereign — Creature — Giant Warrior {5}{R}, 5/5 (Lorwyn, #5593):
//
//	"Other Giant creatures you control get +2/+2 and have trample."
//
// A lord (tribal.go): "other" leaves the source out and "you control" is
// printed, so only your other Giant creatures get it. The boost is layer 7c
// (CR 613.4c) and the trample layer 6 (CR 613.1f); a changeling counts
// (CR 702.73a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "c9abbb90-e3ae-4e7c-87ca-06ebebf65edb",
		Name:         "Sunrise Sovereign",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			TribalAnthem(TribeFilter{Tribes: []string{"Giant"}, Others: true, YoursOnly: true}, 2, 2),
			TribalKeywordGrant(TribeFilter{Tribes: []string{"Giant"}, Others: true, YoursOnly: true}, "trample"),
		},
	})
}
