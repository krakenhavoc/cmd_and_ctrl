package effects

// Elfhame Wurm — Creature — Wurm {4}{G}, 5/4 (Dominaria United, S82):
//
//	"Vigilance, trample"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9a74a1dd-c36f-4900-b835-fc39dd725bde",
		Name:            "Elfhame Wurm",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"vigilance", "trample"},
	})
}
