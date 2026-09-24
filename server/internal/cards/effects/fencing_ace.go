package effects

// Fencing Ace — "Double strike."
//
// Minimal double-strike creature — 1/1 for 1W. Participates in
// BOTH combat damage substeps (CR 510.4), dealing its printed
// power in each.
func init() {
	Register(Spec{
		OracleID:        "2f810936-2ba6-4c2b-84a0-ff4c1deb026b",
		Name:            "Fencing Ace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
