package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrill of the Hunt — Instant {G} (Time Spiral, #5596):
//
//	"Target creature gets +1/+2 until end of turn.
//	 Flashback {W} (You may cast this card from your graveyard for its
//	 flashback cost. Then exile it.)"
//
// A +1/+2 trick with flashback (CR 702.34a) for {W} from the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "7b1f0b78-806f-4a7d-809b-d2bc60cf1b01",
		Name:             "Thrill of the Hunt",
		Completeness:     CompletenessFull,
		Targets:          TargetCreature("target creature"),
		Purpose:          ForTargets(BenefitUntilEOT(0, 1, 2)),
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{W}")},
		OnResolve:        boostTheTargetUntilEOT(1, 2, "Thrill of the Hunt — +1/+2 until end of turn"),
	})
}
