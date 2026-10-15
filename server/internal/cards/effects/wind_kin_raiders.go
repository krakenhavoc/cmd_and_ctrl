package effects

// Wind-Kin Raiders — improvise (CR 702.126) with flying.
//
// Improvise is Spec.TapCost (Improvise()); the keywords are declared as
// printed and the engine does the rest. No simplification.
func init() {
	Register(Spec{
		OracleID:        "74dcff4f-c126-4bbb-b6f0-3f3315bc95fe",
		Name:            "Wind-Kin Raiders",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		TapCost:         Improvise(),
	})
}
