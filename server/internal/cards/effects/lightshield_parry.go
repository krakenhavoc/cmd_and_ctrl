package effects

// Lightshield Parry — Instant {W} (Aetherdrift, #4464):
//
//	"Target creature gets +2/+2 until end of turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// Artful Maneuver's pump without rebound, with cycling (CR 702.29a). One
// target, so a creature gone by resolution counters the spell
// (CR 608.2b); the pump ends at cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ecef9e3d-dce8-405f-bba3-14732444258f",
		Name:         "Lightshield Parry",
		Purpose:      ForTargets(BenefitUntilEOT(0, 2, 2)),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    boostTheTargetUntilEOT(2, 2, "Lightshield Parry — +2/+2"),
		Activated:    []ActivatedAbility{Cycling("{2}")},
	})
}
