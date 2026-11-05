package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hard-Won Jitte — {1}{R} Artifact — Equipment (Teenage Mutant Ninja
// Turtles, #4457):
//
//	"Equipped creature has double strike.
//	 Equip {2}"
//
// Fireshrieker's text at a cheaper casting cost: a layer-6 grant to the
// equipped creature and the ordinary equip ability (CR 702.6a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ab433d41-e801-405c-81f7-06fa6640797d",
		Name:         "Hard-Won Jitte",
		Completeness: CompletenessFull,
		Static: []game.StaticAbility{
			GrantToAttached("double strike"),
		},
		Activated: []ActivatedAbility{
			EquipAbility("{2}"),
		},
	})
}
