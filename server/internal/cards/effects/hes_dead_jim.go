package effects

// He's Dead, Jim — Instant {3}{B} (Star Trek, #4973 batch 12):
//
//	"Destroy target creature. Surveil 1."
//
// Hour of Defeat's text exactly, so it is Hour of Defeat's body: the
// surveil happens whether or not the creature was destroyed, and an
// illegal lone target means the spell does not resolve (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ee7150d1-8b38-4387-9b9c-a40180430170",
		Name:         "He's Dead, Jim",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    destroyTargetThenSurveil(1),
	})
}
