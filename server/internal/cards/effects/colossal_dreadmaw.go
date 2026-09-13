package effects

// Colossal Dreadmaw — "Trample."
//
// Canonical vanilla-trample fatty. 6/6 for 4GG; the printed
// keyword feeds the combat engine's single-blocker overflow math
// + the CR 510.1c damage-assignment prompt's `allow_trample`
// flag for multi-blocker scenarios.
func init() {
	Register(Spec{
		OracleID:        "08c7db90-c0cf-4482-b7ee-bb033e5996d2",
		Name:            "Colossal Dreadmaw",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
