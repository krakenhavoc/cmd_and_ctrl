package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Steppe Glider — Creature — Elemental {4}{W}, 2/4 (Game Night 2019, #5515):
//
//	"Flying, vigilance
//	 {1}{W}: Target creature with a +1/+1 counter on it gains flying and
//	 vigilance until end of turn."
//
// The target must carry a +1/+1 counter, judged at announce and again as the
// ability resolves (CR 608.2b), so a creature that lost its counter in
// response is not granted anything. Both keywords go on together as one
// effect (Elspeth, Knight-Errant's grant with no boost).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "85eddd5e-1117-4d58-ac47-ba651050550f",
		Name:            "Steppe Glider",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{W}: Target creature with a +1/+1 counter on it gains flying and vigilance until end of turn.",
			Cost:    effects.ManaCost("{1}{W}"),
			Targets: effects.TargetCreature("target creature with a +1/+1 counter on it", effects.CreatureWithCounter(game.CounterPlusOne)),
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "flying", "vigilance")), game.AnswerCombatGrant),
			Effect:  effects.PumpAndGrantFirstTarget(0, 0, "Steppe Glider — flying and vigilance", "flying", "vigilance"),
		}},
	})
}
