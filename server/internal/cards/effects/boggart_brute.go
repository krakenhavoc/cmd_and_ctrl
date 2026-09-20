package effects

// Boggart Brute — "Menace."
//
// Simple vanilla menace creature — 3/2 for 2R. Covers the S18
// menace keyword. Menace is a block COUNT (CR 702.111b), so
// DeclareBlockers REFUSES a lone blocker against this attacker with a
// too_few_blockers reason and stores nothing; blocking it takes two
// creatures declared together (CR 509.1b, #750).
//
// Swapped in for Dreg Mangler from the ADR 0014 §2.8 card list
// after oracle-text review showed Dreg Mangler has haste +
// scavenge, NOT menace. Boggart Brute carries only menace, which
// is what the plan needed for the twelfth keyword coverage slot.
func init() {
	Register(Spec{
		OracleID:        "880793a2-84b0-4ae1-9bb8-6e942e3dc723",
		Name:            "Boggart Brute",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
