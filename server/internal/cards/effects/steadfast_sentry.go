package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Steadfast Sentry — {2}{W} Creature — Human Soldier 3/2 (Foundations
// Jumpstart):
//
//	"Vigilance
//	 When this creature dies, put a +1/+1 counter on target creature you
//	 control."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0f676ab-f859-4d4f-abe0-d033ffd3c396",
		Name:            "Steadfast Sentry",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisDies("Steadfast Sentry — put a +1/+1 counter on target creature you control",
				rfCreatureFCountersOnTarget(1)),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
