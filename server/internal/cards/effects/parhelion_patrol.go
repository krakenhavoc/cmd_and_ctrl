package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Parhelion Patrol — Creature — Human Knight {3}{W}, 2/3 (Foundations
// Jumpstart):
//
//	"Flying
//	 Vigilance
//	 Mentor (Whenever this creature attacks, put a +1/+1 counter on
//	 target attacking creature with lesser power.)"
//
// Mentor is the shared Mentor() trigger (CR 702.134), as Blade Instructor's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d4891bfc-12b7-4412-be11-d02e6d014463",
		Name:            "Parhelion Patrol",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Triggered:       []game.TriggeredAbility{Mentor("Parhelion Patrol — mentor")},
	})
}
