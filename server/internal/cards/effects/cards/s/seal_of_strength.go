package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Seal of Strength — Enchantment {G}
// (Nemesis, #5618):
//
//	"Sacrifice this enchantment: Target creature gets +3/+3 until end of turn."
//
// Fires of Yavimaya's sacrifice row at +3/+3: the sacrifice is a cost,
// and the boost ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e41a68b3-e1cb-4f51-bd54-68882d2cc015",
		Name:         "Seal of Strength",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this enchantment: Target creature gets +3/+3 until end of turn.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Purpose: game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 3, 3))},
			Effect:  effects.BoostTargetUntilEOT("Seal of Strength — +3/+3", 3, 3),
		}},
	})
}
