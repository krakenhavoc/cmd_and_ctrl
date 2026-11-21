package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Rouse — Instant {1}{B} (Mercadian Masques, #5619):
//
//	"If you control a Swamp, you may pay 2 life rather than pay this
//	 spell's mana cost.
//	 Target creature gets +2/+0 until end of turn."
//
// Snuff Out's alternative cost at 2 life (PayLifeInstead with
// ControlsA("Swamp"), checked as the spell is cast) on the shared
// boostTheTargetUntilEOT body. The life is a cost, paid with the spell
// on the stack (CR 601.2h), so it is not given back if Rouse is
// countered. The boost ends in the cleanup step (CR 514.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "475099d9-4743-420a-ada8-3932f5ef9e82",
		Name:         "Rouse",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
		Targets:      effects.TargetCreature("target creature"),
		AlternativeCosts: []game.AlternativeCost{
			effects.PayLifeInstead("Pay 2 life (you control a Swamp)", 2, effects.ControlsA("Swamp")),
		},
		OnResolve: effects.BoostTheTargetUntilEOT(2, 0, "Rouse — +2/+0"),
	})
}
