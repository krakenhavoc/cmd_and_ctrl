package effects

// Fencing Ace — "Double strike."
//
// Minimal double-strike creature — 1/1 for 1W. Participates in
// BOTH the first-strike substep (CR 510.2) and the regular
// substep (CR 510.3), dealing its printed power in each.
func init() {
	Register(Spec{
		OracleID:        "2f810936-2ba6-4c2b-84a0-ff4c1deb026b",
		Name:            "Fencing Ace",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"double strike"},
	})
}
