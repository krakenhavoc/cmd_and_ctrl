package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	pump := effects.MrdB5TowerRow("Target creature gets +6/+6 until end of turn.",
		game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 6, 6))},
		effects.BoostTargetUntilEOT("Tower of Champions — +6/+6 until end of turn", 6, 6))
	pump.Targets = effects.TargetCreature("target creature")
	effects.Register(effects.Spec{
		OracleID:     "f7880784-e1dc-4250-812e-eda41fe91e36",
		Name:         "Tower of Champions",
		Completeness: effects.CompletenessFull,
		Activated:    []effects.ActivatedAbility{pump},
	})
}
