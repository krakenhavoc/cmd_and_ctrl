package effects

// Bastion Inventor — improvise (CR 702.126) with hexproof.
//
// Improvise is Spec.TapCost (Improvise()); the keywords are declared as
// printed and the engine does the rest. No simplification.
func init() {
	Register(Spec{
		OracleID:        "900ae045-0160-4e85-b283-e35399c2acec",
		Name:            "Bastion Inventor",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		TapCost:         Improvise(),
	})
}
