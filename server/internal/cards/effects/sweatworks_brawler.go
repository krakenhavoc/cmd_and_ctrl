package effects

// Sweatworks Brawler — improvise (CR 702.126) with menace.
//
// Improvise is Spec.TapCost (Improvise()); the keywords are declared as
// printed and the engine does the rest. No simplification.
func init() {
	Register(Spec{
		OracleID:        "9bb1de3f-b6af-4ee2-b896-a05aae50e238",
		Name:            "Sweatworks Brawler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
		TapCost:         Improvise(),
	})
}
