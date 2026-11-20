package effects

// Sedge Scorpion — Creature — Scorpion {G}, 1/1 (Theros, #5561):
//
//	"Deathtouch (Any amount of damage this deals to a creature is enough to destroy it.)"
//
// Printed keywords only; the engine reads them off the spec.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "fbadf7a3-0d54-40fe-a763-f641dd448e56",
		Name:            "Sedge Scorpion",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
	})
}
