package effects

// Aerial Responder — Creature — Dwarf Soldier {1}{W}{W}, 2/3 (Kaladesh, #5539):
//
//	"Flying, vigilance, lifelink"
//
// Printed keywords only; the engine reads each one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "69b1b0ec-9db0-48d1-a7b5-71281aca16fe",
		Name:            "Aerial Responder",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance", "lifelink"},
	})
}
