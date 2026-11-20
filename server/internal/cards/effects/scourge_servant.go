package effects

// Scourge Servant — Creature — Phyrexian Zombie {4}{B}, 3/3 (Mirrodin Besieged,
// #5576):
//
//	"Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)"
//
// Infect is the engine's damage tail (CR 702.90b): -1/-1 counters to
// creatures, poison counters to players.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "561d514d-060e-491f-a703-8a7f9da5204d",
		Name:            "Scourge Servant",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect"},
	})
}
