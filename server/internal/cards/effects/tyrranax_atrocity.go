package effects

// Tyrranax Atrocity — Creature — Phyrexian Dinosaur {3}{G}{G}, 4/4 (Phyrexia: All Will Be One, #5479):
//
//	"Haste
//	 Toxic 3"
//
// Every printed keyword is read through the layered view. Toxic is the
// engine's damage tail (infect_wither_toxic.go, CR 702.164).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "08261b68-b261-4e47-a085-c489ba394108",
		Name:            "Tyrranax Atrocity",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste", "toxic 3"},
	})
}
