package effects

// Phytoburst — Sorcery {1}{G}:
//
//	"Target creature gets +5/+5 until end of turn."
//
// The shared boostTheTargetUntilEOT body at +5/+5, ending at cleanup
// (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d68d175f-f59c-4995-84d4-fc593280194b",
		Name:         "Phytoburst",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		Purpose:      ForTargets(BenefitUntilEOT(0, 5, 5)),
		OnResolve:    boostTheTargetUntilEOT(5, 5, "Phytoburst — +5/+5 until end of turn"),
	})
}
