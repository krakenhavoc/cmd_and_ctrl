package effects

// Feral Abomination — Creature — Thrull {5}{B}, 5/5 (Dominaria, #5530):
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)"
//
// One printed keyword, read by the state-based check after it deals damage (CR 702.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1bf06342-0d76-433b-9912-8af03b7e3bc3",
		Name:            "Feral Abomination",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
