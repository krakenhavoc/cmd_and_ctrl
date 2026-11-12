package effects

// Dancing Scimitar — Artifact Creature — Spirit {4}, 1/5 (Arabian
// Nights, #5641):
//
//	"Flying (This creature can't be blocked except by creatures with
//	 flying or reach.)"
//
// Flying is an engine keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "82b13601-d460-45c5-94a1-07e146d463a9",
		Name:            "Dancing Scimitar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
	})
}
