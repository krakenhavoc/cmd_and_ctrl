package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Tower of Champions — Artifact {4} (Mirrodin, #5606):
//
//	"{8}, {T}: Target creature gets +6/+6 until end of turn."
//
// A layer 7c boost on the target until cleanup (CR 514.2); a target
// that is gone or no longer a creature as the ability resolves gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	pump := mrdB5TowerRow("Target creature gets +6/+6 until end of turn.",
		game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(BenefitUntilEOT(0, 6, 6))},
		boostTargetUntilEOT("Tower of Champions — +6/+6 until end of turn", 6, 6))
	pump.Targets = TargetCreature("target creature")
	Register(Spec{
		OracleID:     "f7880784-e1dc-4250-812e-eda41fe91e36",
		Name:         "Tower of Champions",
		Completeness: CompletenessFull,
		Activated:    []ActivatedAbility{pump},
	})
}
