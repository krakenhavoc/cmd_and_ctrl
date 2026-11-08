package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Felling Blow — Sorcery {2}{G} (Foundations, #4465):
//
//	"Put a +1/+1 counter on target creature you control. Then that creature
//	 deals damage equal to its power to target creature an opponent
//	 controls."
//
// Knockout Maneuver's text and body (counterOnYourCreatureThenBite). Two
// target clauses, read positionally: slot 0 gets the counter and then deals
// the damage, slot 1 takes it. The power read includes the new counter. A
// creature of yours that left in response gets no counter and deals
// nothing; a recipient that left takes nothing, but your creature still gets
// its counter (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fbfa19ca-98bb-4c7d-a04e-920a8aafb031",
		Name:         "Felling Blow",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove}, // one-sided fight, as Bite Down (#4734)
		Targets: Clauses(
			TargetCreature("target creature you control", YouControl()),
			TargetCreature("target creature an opponent controls", OpponentControls()),
		),
		OnResolve: counterOnYourCreatureThenBite(),
	})
}
