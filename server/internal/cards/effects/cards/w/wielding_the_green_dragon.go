package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wielding the Green Dragon — Sorcery {1}{G} (Portal Three Kingdoms, #5622):
//
//	"Target creature gets +4/+4 until end of turn."
//
// Monstrous Growth's body (plusFourPlusFourToTheTarget). The pump ends at
// cleanup (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "22fa9020-783e-4292-857e-9d45a3458d71",
		Name:         "Wielding the Green Dragon",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.PlusFourPlusFourToTheTarget("Wielding the Green Dragon"),
	})
}
