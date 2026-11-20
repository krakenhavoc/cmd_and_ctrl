package effects

// Cystbearer — Creature — Phyrexian Beast {2}{G}, 2/3 (Scars of Mirrodin, #5577)
//
//	"Infect (This creature deals damage to creatures in the form of -1/-1
//	 counters and to players in the form of poison counters.)"
//
// A keyword body: infect is the engine's damage tail (CR 702.90).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "81704164-bad9-4e52-9b69-3d49e92fe284",
		Name:            "Cystbearer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
	})
}
