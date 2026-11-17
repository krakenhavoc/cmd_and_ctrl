package effects

// Sheoldred's Headcleaver — Creature — Phyrexian Warrior {3}{B}, 2/4 (Phyrexia: All Will Be One, #5479):
//
//	"Menace
//	 Toxic 2"
//
// Every printed keyword is read through the layered view. Toxic is the
// engine's damage tail (infect_wither_toxic.go, CR 702.164).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "a7df9a4e-d556-4345-9884-0724a3f9d5f8",
		Name:            "Sheoldred's Headcleaver",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "toxic 2"},
	})
}
