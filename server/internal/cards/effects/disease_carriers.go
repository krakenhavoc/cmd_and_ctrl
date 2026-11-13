package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Disease Carriers — Creature — Rat {2}{B}{B}, 2/2 (Urza's Destiny,
// #5621):
//
//	"When this creature dies, target creature gets -2/-2 until end of
//	 turn."
//
// The shared udsB2WhenThisDiesTargetCreatureGets at -2/-2, declared a
// harm to its target. The trigger is mandatory, so with no other
// creature on the battlefield one of its controller's own is the
// target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ccdd6376-9b74-4d13-9f80-87d0bd3d4a2a",
		Name:         "Disease Carriers",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			udsB2WhenThisDiesTargetCreatureGets("Disease Carriers", -2, -2),
		},
	})
}
