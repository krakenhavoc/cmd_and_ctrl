package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ursapine — Creature — Beast {3}{G}{G}, 3/3 (Ravnica: City of Guilds, S82 OLD
// batch 575, #5600):
//
//	"{G}: Target creature gets +1/+1 until end of turn."
//
// A +1/+1 pump of the creature targeted as the ability goes on the stack,
// applied as it resolves if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fcf92abb-6c06-42d2-a3d1-f44ec997d56d",
		Name:         "Ursapine",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{G}: Target creature gets +1/+1 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)), game.AnswerPump),
			Cost:    effects.ManaCost("{G}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Ursapine — +1/+1", 1, 1),
		}},
	})
}
