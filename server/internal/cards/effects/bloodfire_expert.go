package effects

// Bloodfire Expert — Creature — Efreet Monk {2}{R}, 3/1 (Foundations
// Jumpstart):
//
//	"Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Prowess is derived by the engine from the keyword (game/prowess.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "341cf654-faf7-4db9-bd9a-70783f1ccff0",
		Name:            "Bloodfire Expert",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"prowess"},
	})
}
