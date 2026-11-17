package effects

// Soltari Priest — Creature — Soltari Cleric {W}{W}, 2/1
// (Tempest, #5628):
//
//	"Protection from red
//	 Shadow (This creature can block or be blocked by only creatures
//	 with shadow.)"
//
// Soltari Monk's shape with red: a red creature can't block it
// (CR 702.16f), red spells and abilities can't target it (CR 702.16b),
// and damage from red sources is prevented (CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f8c9cb9f-3a37-4004-9cde-f51ca5579e1a",
		Name:            "Soltari Priest",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"protection from red", "shadow"},
	})
}
