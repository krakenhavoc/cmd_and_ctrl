package effects

// Bilbo's Deadly Slice — Instant {1}{B}{B} (The Hobbit, HOB):
//
//	"Destroy target creature."
//
// Murder under another name (#4454): same Spec, new oracle id.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "4c054431-3667-4e28-9d68-03b43447c35b",
		Name:         "Bilbo's Deadly Slice",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    destroyTheTargetPermanent,
	})
}
