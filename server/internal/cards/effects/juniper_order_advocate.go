package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Juniper Order Advocate — Creature — Human Knight {2}{W}, 1/2
// (Alliances, #5634):
//
//	"As long as this creature is untapped, green creatures you control
//	 get +1/+1."
//
// A layer 7c anthem (CR 613.4c) over the green creatures its controller
// controls, switched off while the Advocate is tapped. A tap or untap
// invalidates the layers by itself; an attack taps the Advocate without a
// tap event (CR 508.1f), so the static also declares that it reads
// attacking status, which makes the attack declaration invalidate them.
// Built by creaturesOfColorYouControlGet (color_anthems.go).
//
// No simplification.
func init() {
	anthem := creaturesOfColorYouControlGet("G", 1, 1, func(source *game.Card) bool { return !source.Tapped },
		"As long as this creature is untapped, green creatures you control get +1/+1.")
	anthem.DependsOnAttackingStatus = true
	Register(Spec{
		OracleID:     "ded2fc21-1a69-48f8-865a-467e21b2758b",
		Name:         "Juniper Order Advocate",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{anthem},
	})
}
