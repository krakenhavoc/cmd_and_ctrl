package effects

// Silent Dart — Artifact {1} (Guilds of Ravnica, #5525):
//
//	"{4}, {T}, Sacrifice this artifact: It deals 3 damage to target creature."
//
// Bear Trap's ability at {4}: the tap and the sacrifice are paid at
// announcement, so the Dart is gone before the damage resolves and the
// damage comes from its last-known information (CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "912bc348-9ba3-4001-90a8-104d2bb17527",
		Name:         "Silent Dart",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{4}, {T}, Sacrifice this artifact: It deals 3 damage to target creature.",
			Cost:    Plus(ManaCost("{4}"), TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  sourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
