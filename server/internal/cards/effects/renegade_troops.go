package effects

// Renegade Troops — Creature — Human Soldier {4}{R}, 4/2
// (Portal Three Kingdoms, #5622):
//
//	"Haste"
//
// Haste is an engine keyword carried as printed data: it can attack and
// tap the turn it arrives (CR 702.10b, CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "83a81308-4a15-49cd-b2f6-51431cca79e9",
		Name:            "Renegade Troops",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
	})
}
