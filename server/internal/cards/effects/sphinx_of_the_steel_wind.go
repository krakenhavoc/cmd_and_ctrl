package effects

// Sphinx of the Steel Wind — Artifact Creature — Sphinx {5}{W}{U}{B}, 6/6
// (Mystery Booster 2):
//
//	"Flying, first strike, vigilance, lifelink, protection from red and from
//	 green"
//
// Four combat keywords the engine honours, and protection from two colours
// (CR 702.16), declared as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1c96772b-463c-4655-836b-cf6e9dc6319f",
		Name:            "Sphinx of the Steel Wind",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "first strike", "vigilance", "lifelink", "protection from red", "protection from green"},
	})
}
