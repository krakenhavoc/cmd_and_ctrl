package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earth Village Ruffians — Creature — Human Soldier Rogue {2}{B/G}, 3/1
// (Avatar: The Last Airbender, #4459):
//
//	"When this creature dies, earthbend 2. (Target land you control
//	 becomes a 0/0 creature with haste that's still a land. Put two +1/+1
//	 counters on it. When it dies or is exiled, return it to the
//	 battlefield tapped.)"
//
// A dies trigger whose whole effect is the earthbend keyword action, its
// target the keyword's own clause chosen as the trigger goes on the stack
// (CR 603.3d). A land that is no longer a legal target as it resolves
// takes the ability with it (CR 608.2b, the ruling).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0380e99c-18f9-4472-8223-bb208e5f0963",
		Name:         "Earth Village Ruffians",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Earth Village Ruffians — earthbend 2", avB16Earthbend(2)), EarthbendTargets()),
		},
	})
}
