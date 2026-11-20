package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Lose Hope — Instant {B} (Fifth Dawn, #5604):
//
//	"Target creature gets -1/-1 until end of turn. Scry 2. (Look at the
//	 top two cards of your library, then put any number of them on the
//	 bottom and the rest on top in any order.)"
//
// Disfigure's -N/-N at one, then the scry (dn5B2BoostTheTargetThenScry2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4f6dbd90-f7fe-4adf-a00f-27573c0bd5c4",
		Name:         "Lose Hope",
		Purpose:      WithRiders(ForTargets(HarmUntilEOT(0, -1, -1)), game.Purpose{Scry: 2}),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    dn5B2BoostTheTargetThenScry2(-1, -1, "Lose Hope — -1/-1 until end of turn"),
	})
}
