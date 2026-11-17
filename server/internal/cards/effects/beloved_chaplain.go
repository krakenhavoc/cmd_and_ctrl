package effects

// Beloved Chaplain — Creature — Human Cleric {1}{W}, 1/1 (Odyssey, #5613):
//
//	"Protection from creatures"
//
// Protection from a card type (CR 702.16a): no creature can block it
// (CR 702.16f), damage from creatures is prevented (CR 702.16e), and
// abilities of creatures can't target it (CR 702.16b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a07406b7-faa1-4ed3-ab84-314c40f3f7f1",
		Name:            "Beloved Chaplain",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from creatures"},
	})
}
