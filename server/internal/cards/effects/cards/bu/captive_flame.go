package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Captive Flame — Enchantment {2}{R} (Saviors of Kamigawa, #5601):
//
//	"{R}: Target creature gets +1/+0 until end of turn."
//
// One pump per activation, each ending in the cleanup step (CR 514.2); a
// target that left in response is skipped (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fada1102-bbb7-4c90-a72e-6c595c08a55d",
		Name:         "Captive Flame",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}: Target creature gets +1/+0 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0)), game.AnswerPump),
			Cost:    effects.ManaCost("{R}"),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NeoB07TargetGetsUntilEOT("Captive Flame — +1/+0 until end of turn", 1, 0),
		}},
	})
}
