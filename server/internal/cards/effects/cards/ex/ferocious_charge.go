package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ferocious Charge — Instant {2}{G} (Fifth Dawn, #5604):
//
//	"Target creature gets +4/+4 until end of turn. Scry 2. (Look at the
//	 top two cards of your library, then put any number of them on the
//	 bottom and the rest on top in any order.)"
//
// Titan's Strength's shape: the pump, then the scry, which happens
// whenever the spell resolves (dn5B2BoostTheTargetThenScry2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a7796af7-d41c-42a6-9033-a4d79dee84d6",
		Name:         "Ferocious Charge",
		Purpose:      effects.WithRiders(effects.ForTargets(effects.BenefitUntilEOT(0, 4, 4)), game.Purpose{Scry: 2}),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.Dn5B2BoostTheTargetThenScry2(4, 4, "Ferocious Charge — +4/+4 until end of turn"),
	})
}
