package effects

// Tangle Mantis — Creature — Insect {2}{G}{G}, 3/4 (Mirrodin Besieged,
// #5576):
//
//	"Trample"
//
// Trample is a printed keyword read at combat damage (CR 702.19b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d52397bf-e3fb-4894-996d-7c1cc8d92603",
		Name:            "Tangle Mantis",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
