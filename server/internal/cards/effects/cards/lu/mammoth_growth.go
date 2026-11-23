package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mammoth Growth — Instant {2}{G} (Kaldheim, #5505):
//
//	"Target creature gets +4/+4 until end of turn.
//	 Foretell {G}"
//
// Monstrous Growth's body (plusFourPlusFourToTheTarget), plus the
// shared foretell special action (CR 702.143).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "bd433a27-1f3f-4d69-8b7e-97239fdf165b",
		Name:           "Mammoth Growth",
		Completeness:   effects.CompletenessFull,
		Purpose:        effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)),
		Targets:        effects.TargetCreature("target creature"),
		SpecialActions: []game.SpecialAction{effects.Foretell("{G}")},
		OnResolve:      effects.PlusFourPlusFourToTheTarget("Mammoth Growth"),
	})
}
