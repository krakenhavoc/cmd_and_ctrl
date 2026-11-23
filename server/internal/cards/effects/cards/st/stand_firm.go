package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stand Firm — Instant {W} (Fifth Dawn, #5604):
//
//	"Target creature gets +1/+1 until end of turn. Scry 2. (Look at the
//	 top two cards of your library, then put any number of them on the
//	 bottom and the rest on top in any order.)"
//
// Titan's Strength's shape: the pump, then the scry, which happens
// whenever the spell resolves (dn5B2BoostTheTargetThenScry2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c0406e70-8131-4e13-b1d5-4e943ad296b8",
		Name:         "Stand Firm",
		Purpose:      effects.WithRiders(effects.ForTargets(effects.BenefitUntilEOT(0, 1, 1)), game.Purpose{Scry: 2}),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.Dn5B2BoostTheTargetThenScry2(1, 1, "Stand Firm — +1/+1 until end of turn"),
	})
}
