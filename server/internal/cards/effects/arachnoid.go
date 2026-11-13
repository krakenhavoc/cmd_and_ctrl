package effects

// Arachnoid — Artifact Creature — Spider {6}, 2/6 (Fifth Dawn, #5604):
//
//	"Reach (This creature can block creatures with flying.)"
//
// The keyword rides PrintedKeywords; the combat engine does the rest.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7116db6d-82ea-4675-a0cd-73aaf94be43f",
		Name:            "Arachnoid",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach"},
	})
}
