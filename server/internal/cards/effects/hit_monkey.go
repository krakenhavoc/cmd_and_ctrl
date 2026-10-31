package effects

// Hit-Monkey — Legendary Creature — Monkey Assassin {3}{G}, 3/3
// (Marvel Super Heroes Commander, #4455):
//
//	"This spell can't be countered.
//	 Reach, vigilance, deathtouch, hexproof, haste"
//
// The first line is a statement about the spell on the stack
// (CantBeCountered, CR 101.2); the rest are printed keywords.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e27d1993-8d83-449d-8a64-8b66086db995",
		Name:            "Hit-Monkey",
		Completeness:    CompletenessFull,
		CantBeCountered: true,
		PrintedKeywords: []string{"reach", "vigilance", "deathtouch", "hexproof", "haste"},
	})
}
