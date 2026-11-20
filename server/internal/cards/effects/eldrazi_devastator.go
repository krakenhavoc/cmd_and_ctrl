package effects

// Eldrazi Devastator — Creature — Eldrazi {8}, 8/9 (Battle for
// Zendikar, #5547):
//
//	"Trample"
//
// Trample is an engine keyword carried as printed data (CR 702.19b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "bc08ac66-969c-4905-a2d6-16ff565be475",
		Name:            "Eldrazi Devastator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
