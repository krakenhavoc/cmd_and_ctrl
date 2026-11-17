package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Astral Steel — Instant {2}{W} (Scourge):
//
//	"Target creature gets +1/+2 until end of turn.
//	 Storm (When you cast this spell, copy it for each spell cast before
//	 it this turn. You may choose new targets for the copies.)"
//
// Giant Growth's body for +1/+2, plus the storm keyword (CR 702.40a). Each
// copy is its own pump and may be pointed at a different creature (CR
// 707.10c); copies left on the same creature add up.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef8ac291-3789-4147-a11f-c62fd8566dbe",
		Name:         "Astral Steel",
		Completeness: CompletenessFull,
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 2)),
		Targets:      TargetCreature("target creature"),
		Triggered:    []game.TriggeredAbility{Storm()},
		OnResolve:    boostTheTargetUntilEOT(1, 2, "Astral Steel — +1/+2 until end of turn"),
	})
}
