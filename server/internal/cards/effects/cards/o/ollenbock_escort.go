package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ollenbock Escort — Creature — Human Cleric {W}, 1/1 (Crimson Vow):
//
//	"Vigilance
//	 Sacrifice this creature: Target creature you control with a +1/+1
//	 counter on it gains lifelink and indestructible until end of turn."
//
// Vigilance rides PrintedKeywords. The sacrifice is the cost, so the Escort
// is never its own target; the target must carry a +1/+1 counter when chosen
// and when the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a8fa68b-8c87-4ecd-ac29-50fd90c62159",
		Name:            "Ollenbock Escort",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Target creature you control with a +1/+1 counter on it gains lifelink and indestructible until end of turn.",
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature you control with a +1/+1 counter on it", effects.YouControl(), effects.VowB09HasPlusOneCounter()),
			Purpose: effects.WithAnswers(effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "lifelink", "indestructible")), game.AnswerSurvive|game.AnswerCombatGrant),
			Effect:  effects.VowB09GrantUntilEOTOnTarget("Ollenbock Escort — lifelink and indestructible until end of turn", "lifelink", "indestructible"),
		}},
	})
}
