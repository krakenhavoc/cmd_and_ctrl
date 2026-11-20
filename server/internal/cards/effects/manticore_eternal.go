package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Manticore Eternal — Creature — Zombie Manticore {3}{R}{R}, 5/4
// (Hour of Devastation, #5534):
//
//	"Afflict 3 (Whenever this creature becomes blocked, defending player
//	 loses 3 life.)
//	 This creature attacks each combat if able."
//
// Afflict is the shared Afflict row (CR 702.130a): when this creature
// becomes blocked, the defending player loses the life, which is life
// loss and not damage.
// The attack requirement is the shared AttacksEachCombat static (CR 508.1d).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d859e9b3-7804-418d-be95-e03a5de290ca",
		Name:         "Manticore Eternal",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Afflict(3, "Manticore Eternal")},
		Static:       []game.StaticAbility{AttacksEachCombat()},
	})
}
