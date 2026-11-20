package effects

// Bassara Tower Archer — Creature — Human Archer {G}{G}, 2/1 (Journey into Nyx, #5558):
//
//	"Reach, hexproof"
//
// Printed keywords only, which the engine enforces.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "39daa265-8d02-4449-9c45-ae8ec2b13e5b",
		Name:            "Bassara Tower Archer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "hexproof"},
	})
}
