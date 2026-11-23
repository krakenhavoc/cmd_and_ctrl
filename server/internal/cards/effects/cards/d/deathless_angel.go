package cardsd

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Deathless Angel — Creature — Angel {4}{W}{W}, 5/7 (Rise of the Eldrazi,
// #5581):
//
//	"Flying
//	 {W}{W}: Target creature gains indestructible until end of turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "6098ba4b-ebfe-48da-a33d-b94c021ac559",
		Name:            "Deathless Angel",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{W}{W}: Target creature gains indestructible until end of turn.",
			Cost:    effects.ManaCost("{W}{W}"),
			Targets: effects.TargetCreature("target creature"),
			Purpose: game.Purpose{Answers: game.AnswerSurvive, Targets: game.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible"))},
			Effect:  effects.EbTargetCreatureGainsUntilEOT("indestructible", "Deathless Angel — indestructible until end of turn"),
		}},
	})
}
