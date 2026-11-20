package effects

// Ember Hauler — Creature — Goblin {R}{R}, 2/2 (Magic 2011, #5579):
//
//	"{1}, Sacrifice this creature: It deals 2 damage to any target."
//
// Goblin Legionnaire's damage row at {1}: the sacrifice is part of the
// cost, and the damage is dealt by the Hauler as it last existed on the
// battlefield (CR 113.7a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7d313958-49ca-4592-97c7-20f12c8468a8",
		Name:         "Ember Hauler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{1}, Sacrifice this creature: It deals 2 damage to any target.",
			Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
			Targets: TargetAny(),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  b33DamageChosenTargetFromSource(2),
		}},
	})
}
