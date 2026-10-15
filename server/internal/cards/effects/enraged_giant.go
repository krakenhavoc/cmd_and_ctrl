package effects

// Enraged Giant — improvise (CR 702.126) with trample, haste.
//
// Improvise is Spec.TapCost (Improvise()); the keywords are declared as
// printed and the engine does the rest. No simplification.
func init() {
	Register(Spec{
		OracleID:        "dbf4e56e-57dd-4f41-95a9-8b9b096274ec",
		Name:            "Enraged Giant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "haste"},
		TapCost:         Improvise(),
	})
}
