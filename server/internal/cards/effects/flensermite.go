package effects

// Flensermite — Creature — Phyrexian Gremlin {1}{B}, 1/1 (Mirrodin Besieged,
// #5576):
//
//	"Infect (This creature deals damage to creatures in the form of
//	 -1/-1 counters and to players in the form of poison counters.)
//	 Lifelink (Damage dealt by this creature also causes you to gain
//	 that much life.)"
//
// Infect and lifelink are both printed keywords the engine's damage tail
// reads: the damage becomes -1/-1 counters or poison counters, and it still
// gains its controller that much life (CR 702.90b, 702.15b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "4b63bf36-cd0e-41d8-b0ca-aac459d91f3f",
		Name:            "Flensermite",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"infect", "lifelink"},
	})
}
