package effects

// Voice of Reason — Creature — Angel {3}{W}, 2/2 (Urza's Destiny, #5621):
//
//	"Flying, protection from blue"
//
// Voice of Duty's twin with protection from blue: blue spells and
// abilities from blue sources can't target it (CR 702.16b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "372e2d6e-427c-4152-b936-27dcf536ca84",
		Name:            "Voice of Reason",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "protection from blue"},
	})
}
