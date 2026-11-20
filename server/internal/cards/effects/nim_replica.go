package effects

// Nim Replica — Artifact Creature — Zombie {3}, 3/1 (Mirrodin, #5606):
//
//	"{2}{B}, Sacrifice this creature: Target creature gets -1/-1 until
//	 end of turn."
//
// Phyrexian Denouncer's shrink paid with {2}{B} instead of a tap: the
// Replica is sacrificed as the cost, and the target, checked again as
// the ability resolves (CR 608.2b), gets -1/-1 until cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "68c3aa66-85f2-494c-8fb2-ad0348814506",
		Name:         "Nim Replica",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}, Sacrifice this creature: Target creature gets -1/-1 until end of turn.",
			Purpose: ForTargets(HarmUntilEOT(0, -1, -1)),
			Cost:    Plus(ManaCost("{2}{B}"), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Effect:  boostTargetUntilEOT("Nim Replica — -1/-1", -1, -1),
		}},
	})
}
