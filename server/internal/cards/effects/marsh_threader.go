package effects

// Marsh Threader — Creature — Kor Scout {1}{W}, 2/1 (Worldwake, #5582):
//
//	"Swampwalk (This creature can't be blocked as long as defending player
//	 controls a Swamp.)"
//
// The printed keywords are engine keywords carried as data, as on
// Bog Wraith.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7b2eb6f2-bd13-4eb6-b91a-c0fa8f3d96af",
		Name:            "Marsh Threader",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
	})
}
