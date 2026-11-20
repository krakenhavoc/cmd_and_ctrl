package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Khenra Eternal — Creature — Zombie Jackal Warrior {1}{B}, 2/2
// (Hour of Devastation, #5534):
//
//	"Afflict 1 (Whenever this creature becomes blocked, defending player
//	 loses 1 life.)"
//
// Afflict is the shared Afflict row (CR 702.130a): when this creature
// becomes blocked, the defending player loses the life, which is life
// loss and not damage.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "91b46b22-8062-434b-8432-5ed1062ccaf9",
		Name:         "Khenra Eternal",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Afflict(1, "Khenra Eternal")},
	})
}
