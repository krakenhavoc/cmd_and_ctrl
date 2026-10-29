package effects

// Large Bear — Creature — Bear {3}{B/G}{B/G}, 5/5 (The Hobbit):
//
//	"Reach, trample, haste"
//
// All three keywords ride PrintedKeywords. The hybrid cost is printed
// card data and needs nothing here.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "73346137-3117-4dc3-b5a4-1aca1da54548",
		Name:            "Large Bear",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "trample", "haste"},
	})
}
