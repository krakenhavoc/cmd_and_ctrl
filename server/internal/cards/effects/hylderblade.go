package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hylderblade — Artifact — Equipment {B} (Edge of Eternities):
//
//	"Equipped creature gets +3/+1.
//	 Void — At the beginning of your end step, if a nonland permanent left
//	 the battlefield this turn or a spell was warped this turn, attach this
//	 Equipment to target creature you control.
//	 Equip {4}"
//
// Void is an intervening if (void.go): checked as the end step begins and
// again as the ability resolves. The attach is the equip ability's own
// resolution (AttachSourceToTarget): a target that is no longer legal, or
// an Equipment that has left or come back as a new object, does nothing
// (CR 701.3b, CR 400.7). With no creature of yours the trigger is not put
// on the stack (CR 603.3d).
//
// No simplification.
func init() {
	attach := AtYourEndStepIfVoid("Hylderblade — attach this Equipment to target creature you control",
		AttachSourceToTarget)
	attach.Targets = TargetCreature("target creature you control", YouControl())
	Register(Spec{
		OracleID:     "64ca59b8-6e90-46c8-8047-6172b9c84226",
		Name:         "Hylderblade",
		Completeness: CompletenessFull,
		Static:       []game.StaticAbility{PumpAttached(3, 1)},
		Triggered:    []game.TriggeredAbility{attach},
		Activated: []ActivatedAbility{
			EquipAbility("{4}"),
		},
	})
}
