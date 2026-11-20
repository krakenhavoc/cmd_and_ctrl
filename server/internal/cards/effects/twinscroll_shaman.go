package effects

// Twinscroll Shaman — Creature — Dwarf Shaman {2}{R}, 1/2 (Strixhaven:
// School of Mages, #5502):
//
//	"Double strike"
//
// One printed keyword the combat engine reads.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a2f867cd-410d-4524-8660-38ad142f6b1f",
		Name:            "Twinscroll Shaman",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
