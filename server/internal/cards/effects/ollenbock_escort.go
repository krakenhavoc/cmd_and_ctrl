package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:        "7a8fa68b-8c87-4ecd-ac29-50fd90c62159",
		Name:            "Ollenbock Escort",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance"},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Target creature you control with a +1/+1 counter on it gains lifelink and indestructible until end of turn.",
			Cost:    SacrificeThis(),
			Targets: TargetCreature("target creature you control with a +1/+1 counter on it", YouControl(), vowB09HasPlusOneCounter()),
			Purpose: withAnswers(ForTargets(BenefitUntilEOT(0, 0, 0, "lifelink", "indestructible")), game.AnswerSurvive|game.AnswerCombatGrant),
			Effect:  vowB09GrantUntilEOTOnTarget("Ollenbock Escort — lifelink and indestructible until end of turn", "lifelink", "indestructible"),
		}},
	})
}
