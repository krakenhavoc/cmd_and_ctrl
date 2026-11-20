package effects

// Sewn-Eye Drake — Creature — Zombie Drake {2}{U/R}{B}, 3/1 (Alara
// Reborn, #5586):
//
//	"Flying, haste"
//
// Two printed keywords the engine enforces.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "994cf8bb-c900-448f-b055-ecc2ee6cf9e2",
		Name:            "Sewn-Eye Drake",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste"},
	})
}
