package effects

// Abbey Gargoyles — Creature — Gargoyle {2}{W}{W}{W}, 3/4 (Homelands, #5635):
//
//	"Flying, protection from red"
//
// Both are engine keywords carried as printed data: flying limits its
// blockers (CR 702.9b), and protection from red stops red spells and
// abilities from red sources targeting it and prevents damage from red
// sources (CR 702.16b, CR 702.16e).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "c208b959-d0e4-4a9a-8255-2c7cc7596767",
		Name:            "Abbey Gargoyles",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from red"},
	})
}
