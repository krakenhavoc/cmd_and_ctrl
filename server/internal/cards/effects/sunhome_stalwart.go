package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Sunhome Stalwart — Creature — Human Soldier {1}{W}, 2/2 (Foundations Jumpstart):
//
//	"First strike
//	 Mentor (Whenever this creature attacks, put a +1/+1 counter on
//	 target attacking creature with lesser power.)"
//
// Mentor (CR 702.134) is the shared Mentor() trigger; first strike is an
// engine keyword.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "340d6218-d351-4b2e-80da-7d4afa77190a",
		Name:            "Sunhome Stalwart",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Triggered:       []game.TriggeredAbility{Mentor("Sunhome Stalwart — mentor")},
	})
}
