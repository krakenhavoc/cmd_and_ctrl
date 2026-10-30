package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Foggy Nelson, On Retainer — Legendary Creature — Human Advisor {2}{W},
// 1/3 (Marvel Super Heroes Commander, #4455):
//
//	"Flash
//	 When Foggy Nelson enters, put a +1/+1 counter on another target
//	 creature you control. It gains hexproof until end of turn."
//
// Two instructions on the same target, in order: the counter, then the
// hexproof (CR 608.2c). A target that became illegal gets neither (CR
// 608.2b). Foggy himself is never a legal pick.
//
// No simplification.
func init() {
	enters := WhenThisEnters("Foggy Nelson — +1/+1 counter and hexproof on another creature you control",
		Do(Instruction(putPlusOneCounterOnTheLegalTarget), Instruction(msc10HexproofTheLegalTarget)))
	enters.Targets = Another(TargetCreature("another target creature you control", YouControl()))
	Register(Spec{
		OracleID:        "734fc893-b736-448a-b851-826976031754",
		Name:            "Foggy Nelson, On Retainer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Triggered:       []game.TriggeredAbility{enters},
	})
}
