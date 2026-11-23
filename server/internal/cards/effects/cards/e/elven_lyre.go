package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elven Lyre — Artifact {2} (Fallen Empires, #5637):
//
//	"{1}, {T}, Sacrifice this artifact: Target creature gets +2/+2 until
//	 end of turn."
//
// The Lyre is sacrificed as part of the cost (CR 602.2), so it is gone
// before the pump resolves; a target that left in response gets
// nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7601378d-42cb-4351-8316-8f78a3b49a85",
		Name:         "Elven Lyre",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}, {T}, Sacrifice this artifact: Target creature gets +2/+2 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)), game.AnswerPump),
			Cost:    effects.Plus(effects.ManaCost("{1}"), effects.TapCost(), effects.SacrificeThis()),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PumpAndGrantFirstTarget(2, 2, "Elven Lyre — +2/+2 until end of turn"),
		}},
	})
}
