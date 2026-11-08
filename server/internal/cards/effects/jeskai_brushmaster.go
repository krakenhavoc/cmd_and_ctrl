package effects

// Jeskai Brushmaster — Creature — Orc Monk {1}{U}{R}{W}, 2/4 (Tarkir:
// Dragonstorm, #4463):
//
//	"Double strike
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Two printed keywords, both engine-driven: double strike deals damage in
// both combat damage steps (CR 702.4b), and prowess is the keyword trigger
// (CR 702.108a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "eff690a7-c2c7-4750-894c-0667d3151165",
		Name:            "Jeskai Brushmaster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "prowess"},
	})
}
