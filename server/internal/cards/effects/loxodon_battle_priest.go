package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Loxodon Battle Priest — Creature — Elephant Cleric {4}{W}, 3/5 (Tarkir:
// Dragonstorm, #4463):
//
//	"At the beginning of combat on your turn, put a +1/+1 counter on
//	 another target creature you control."
//
// Luminarch Aspirant's trigger with "another": the Priest can't target
// itself, so with no other creature you control the trigger is removed
// (CR 603.3d). A target that has left by resolution gets nothing (CR
// 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8cf4ddb0-2a83-4d8f-b636-7ba3c81ea4ca",
		Name:         "Loxodon Battle Priest",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Loxodon Battle Priest — put a +1/+1 counter on another target creature you control",
					PlusOneCounterOnChosenTarget),
				Another(TargetCreature("another target creature you control", YouControl()))),
		},
	})
}
