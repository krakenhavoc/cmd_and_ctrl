package effects

// Fanatical Firebrand — Creature — Goblin Pirate {R}, 1/1 (Foundations,
// #4465):
//
//	"Haste
//	 {T}, Sacrifice this creature: It deals 1 damage to any target."
//
// Haste rides PrintedKeywords, so the tap symbol can be paid the turn the
// Goblin arrives. The tap and the sacrifice are both costs (CR 602.2b), so
// the damage comes from the Goblin as it last existed, after it is gone,
// to the target if it is still legal (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d36e11f1-6ab3-4273-8114-a8fbbe21c1c3",
		Name:            "Fanatical Firebrand",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated: []ActivatedAbility{{
			Label:   "{T}, Sacrifice this creature: It deals 1 damage to any target.",
			Cost:    Plus(TapCost(), SacrificeThis()),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 1)),
			Effect:  b33DamageChosenTargetFromSource(1),
		}},
	})
}
