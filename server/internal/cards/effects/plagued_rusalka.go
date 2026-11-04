package effects

// Plagued Rusalka — Creature — Spirit {B}, 1/1 (Jumpstart, #4892):
//
//	"{B}, Sacrifice a creature: Target creature gets -1/-1 until end of
//	 turn."
//
// The sacrifice is a cost, paid at announcement; a target that has left
// by resolution makes the ability do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d6912f52-e478-4cc6-a34c-67e372abc4fa",
		Name:         "Plagued Rusalka",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}, Sacrifice a creature: Target creature gets -1/-1 until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, -1, -1)),
			Cost:    Plus(ManaCost("{B}"), SacrificeACreature()),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Plagued Rusalka — -1/-1", -1, -1),
		}},
	})
}
