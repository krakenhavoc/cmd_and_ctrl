package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pileated Provisioner — Creature — Bird Scout {4}{W}, 3/4 (Bloomburrow,
// #4467):
//
//	"Flying
//	 When this creature enters, put a +1/+1 counter on target creature you
//	 control without flying."
//
// The enters trigger targets as it goes on the stack: a creature you
// control without flying, which the Provisioner itself never is while it
// has flying. With no such creature the trigger is removed (CR 603.3d).
// The target is checked again as it resolves (CR 608.2b), so a creature
// that gained flying in response gets no counter.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0534e20d-8352-4d9f-a617-378070602673",
		Name:            "Pileated Provisioner",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			Targeting(
				WhenThisEnters("Pileated Provisioner — put a +1/+1 counter on target creature you control without flying",
					PlusOneCounterOnTheTarget),
				TargetCreature("target creature you control without flying", YouControl(), WithoutKeyword("flying"))),
		},
	})
}
