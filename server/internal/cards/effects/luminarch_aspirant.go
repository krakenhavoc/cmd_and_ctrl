package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Luminarch Aspirant — Creature — Human Cleric {1}{W}, 1/1:
//
//	"At the beginning of combat on your turn, put a +1/+1 counter on
//	 target creature you control."
//
// A targeted trigger at the beginning of the controller's combat. The
// target is any creature the controller has, the Aspirant included
// (it is not "another"), and the engine re-checks it on resolution
// (CR 608.2b): a creature that left in response is skipped, and the
// trigger does nothing when the controller has no creature at all.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cb9994b9-924b-4e10-9075-9cfbec88f2bf",
		Name:         "Luminarch Aspirant",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Targeting(
				AtBeginningOfYourCombat("Luminarch Aspirant — put a +1/+1 counter on target creature you control",
					PlusOneCounterOnChosenTarget),
				TargetCreature("target creature you control", YouControl())),
		},
	})
}
