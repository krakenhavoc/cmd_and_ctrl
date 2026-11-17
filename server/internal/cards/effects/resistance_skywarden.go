package effects

// Resistance Skywarden — Creature — Ogre Rebel {3}{R}{R}, 5/5 (Phyrexia: All Will Be One, #5479):
//
//	"Reach
//	 Menace"
//
// Every printed keyword is read through the layered view. Toxic is the
// engine's damage tail (infect_wither_toxic.go, CR 702.164).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "419e9ab0-fe79-4a4f-8bcb-fe4e5373e292",
		Name:            "Resistance Skywarden",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"reach", "menace"},
	})
}
