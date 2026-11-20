package effects

// Sphinx of the Steel Wind — Artifact Creature — Sphinx {5}{W}{U}{B},
// 6/6 (Alara Reborn, #5586):
//
//	"Flying, first strike, vigilance, lifelink, protection from red and
//	 from green"
//
// Printed keywords the engine enforces; "protection from red and from
// green" is two protection keywords (CR 702.16a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1c96772b-463c-4655-836b-cf6e9dc6319f",
		Name:         "Sphinx of the Steel Wind",
		Completeness: CompletenessFull,
		PrintedKeywords: []string{
			"flying", "first strike", "vigilance", "lifelink",
			"protection from red", "protection from green",
		},
	})
}
