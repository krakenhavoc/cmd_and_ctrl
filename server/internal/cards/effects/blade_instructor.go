package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Blade Instructor — Creature — Human Soldier {2}{W}, 3/1:
//
//	"Mentor (Whenever this creature attacks, put a +1/+1 counter on
//	 target attacking creature with lesser power.)"
//
// Mentor (CR 702.136) is the shared Mentor() trigger; see
// hammer_dropper.go for what its source-relative clause guarantees.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77cff683-013c-4568-8ae0-f86ef0590fe1",
		Name:         "Blade Instructor",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Mentor("Blade Instructor — mentor")},
	})
}
