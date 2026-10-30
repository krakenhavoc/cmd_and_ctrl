package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Virtuous Variant — Creature — Spider Human Hero {2}{W}, 1/2 (Marvel Super
// Heroes Commander, #4455):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on target creature you
//	 control."
//
// Hunting Moa's counter body (putPlusOneCounterOnEachLegalTarget) on an
// enters trigger whose target is restricted to your own creatures. It can
// target itself.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fa4ed027-6e18-4a87-9845-e13f274d9994",
		Name:            "Virtuous Variant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(WhenThisEnters("Virtuous Variant — put a +1/+1 counter on target creature you control", putPlusOneCounterOnEachLegalTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
