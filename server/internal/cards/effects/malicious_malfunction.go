package effects

// Malicious Malfunction — Sorcery {1}{B}{B}:
//
//	"All creatures get -2/-2 until end of turn. If a creature would die
//	 this turn, exile it instead."
//
// Flaying Tendrils without devoid; see flaying_tendrils.go.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "c7ecaa1a-fbf7-436b-a1c9-d7810b0dc5dc",
		Name:         "Malicious Malfunction",
		Completeness: CompletenessFull,
		OnResolve:    allCreaturesShrinkThenExileIfTheyDie(-2, false),
	})
}
