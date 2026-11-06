package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Earth Kingdom Soldier — Creature — Human Soldier {4}{G/W}, 3/4 (Avatar:
// The Last Airbender, #4459):
//
//	"Vigilance
//	 When this creature enters, put a +1/+1 counter on each of up to two
//	 target creatures you control."
//
// Rishkar's enters trigger narrowed to creatures you control: up to two
// targets chosen as it goes on the stack, which may include the Soldier,
// and a counter on each still legal as it resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4289f2d4-3db2-4dd1-9c18-af3f78309844",
		Name:            "Earth Kingdom Soldier",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Earth Kingdom Soldier — put a +1/+1 counter on each of up to two target creatures you control",
					putPlusOneCounterOnEachLegalTarget),
				TargetCreature("up to two target creatures you control", YouControl()).WithCount(0, 2)),
		},
	})
}
