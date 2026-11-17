package effects

// Ichorspit Basilisk — Creature — Phyrexian Basilisk {2}{G}, 1/3 (Phyrexia: All Will Be One, #5479):
//
//	"Deathtouch
//	 Toxic 1"
//
// Every printed keyword is read through the layered view. Toxic is the
// engine's damage tail (infect_wither_toxic.go, CR 702.164).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5a2f8474-b032-4567-a569-3f9725805f10",
		Name:            "Ichorspit Basilisk",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch", "toxic 1"},
	})
}
