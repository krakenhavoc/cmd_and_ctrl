package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deathless Angel — Creature — Angel {4}{W}{W}, 5/7 (Rise of the Eldrazi,
// #5581):
//
//	"Flying
//	 {W}{W}: Target creature gains indestructible until end of turn."
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6098ba4b-ebfe-48da-a33d-b94c021ac559",
		Name:            "Deathless Angel",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "{W}{W}: Target creature gains indestructible until end of turn.",
			Cost:    ManaCost("{W}{W}"),
			Targets: TargetCreature("target creature"),
			Purpose: game.Purpose{Answers: game.AnswerSurvive, Targets: game.ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible"))},
			Effect:  ebTargetCreatureGainsUntilEOT("indestructible", "Deathless Angel — indestructible until end of turn"),
		}},
	})
}
