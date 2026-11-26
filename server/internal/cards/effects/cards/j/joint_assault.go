package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Joint Assault — {G} Instant: "Target creature gets +2/+2 until end of turn.
// If it's paired with a creature, that creature also gets +2/+2 until end of
// turn."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0f719e56-5203-4138-bbe9-6a37b8ac2512",
		Name:         "Joint Assault",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve:    effects.PumpTargetAndItsSoulbondPartner,
	})
}
