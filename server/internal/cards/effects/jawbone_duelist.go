package effects

// Jawbone Duelist — Creature — Phyrexian Soldier {1}{W}, 1/1 (Phyrexia: All Will Be One, #5479):
//
//	"Double strike
//	 Toxic 1"
//
// Every printed keyword is read through the layered view. Toxic is the
// engine's damage tail (infect_wither_toxic.go, CR 702.164).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "14af3ec5-d90a-4560-a70a-8645765f1253",
		Name:            "Jawbone Duelist",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike", "toxic 1"},
	})
}
