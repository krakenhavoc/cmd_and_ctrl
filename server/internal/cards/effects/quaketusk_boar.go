package effects

// Quaketusk Boar — Creature — Boar {3}{R}{R}, 4/3 (Bloomburrow, #4467):
//
//	"Reach, trample, haste"
//
// Three printed keywords, nothing else.
func init() {
	Register(Spec{
		OracleID:        "6fbaa853-48d4-4a3b-b22d-27415cc266d7",
		Name:            "Quaketusk Boar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample", "haste"},
	})
}
