package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nezumi Linkbreaker — Creature — Rat Warlock {B}, 1/1 (Outlaws of
// Thunder Junction, #4468):
//
//	"When this creature dies, create a 1/1 red Mercenary creature token
//	 with '{T}: Target creature you control gets +1/+0 until end of
//	 turn. Activate only as a sorcery.'"
//
// A dies trigger (CR 603.10a) whose token is the set's Mercenary
// template (MercenaryToken), created under the trigger's controller.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5d24fb6b-7176-407c-b917-6c88a6da40df",
		Name:         "Nezumi Linkbreaker",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisDies("Nezumi Linkbreaker — create a 1/1 red Mercenary", createAMercenary()),
		},
	})
}
