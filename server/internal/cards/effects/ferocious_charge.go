package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "a7796af7-d41c-42a6-9033-a4d79dee84d6",
		Name:         "Ferocious Charge",
		Purpose:      WithRiders(ForTargets(BenefitUntilEOT(0, 4, 4)), game.Purpose{Scry: 2}),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    dn5B2BoostTheTargetThenScry2(4, 4, "Ferocious Charge — +4/+4 until end of turn"),
	})
}
