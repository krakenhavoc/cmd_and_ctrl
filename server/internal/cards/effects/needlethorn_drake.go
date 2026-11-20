package effects

// Needlethorn Drake — Creature — Drake {G}{U}, 1/1 (Strixhaven: School of
// Mages, #5502):
//
//	"Flying, deathtouch"
//
// Two printed keywords the combat engine reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6edacc06-9523-4ccf-a4c2-2ac4af386b8a",
		Name:            "Needlethorn Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "deathtouch"},
	})
}
