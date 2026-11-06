package effects

// Dawn's Light Archer — {2}{G} Creature — Elf Archer 4/2 (Lorwyn
// Eclipsed, #4458):
//
//	"Flash
//	 Reach"
//
// Both keywords ride PrintedKeywords, as Long-Bodied Grey Dog's do: flash
// lets it be cast any time you could cast an instant, and reach lets it
// block creatures with flying.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1f2d1de6-42e0-4f45-b43d-ce4980146469",
		Name:            "Dawn's Light Archer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "reach"},
	})
}
