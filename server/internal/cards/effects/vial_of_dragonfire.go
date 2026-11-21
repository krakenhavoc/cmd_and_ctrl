package effects

// Vial of Dragonfire — Artifact {2} (Commander Masters):
//
//	"{2}, {T}, Sacrifice this artifact: It deals 2 damage to target
//	 creature."
//
// Barrels of Blasting Jelly's second ability at {2} and 2 damage: the
// damage is dealt by the Vial as it last was (CR 608.2h), and a target that
// is gone or illegal at resolution takes nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "17be934e-1e33-4b95-a984-242fa018e280",
		Name:         "Vial of Dragonfire",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}, {T}, Sacrifice this artifact: It deals 2 damage to target creature.",
			Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 2)),
			Effect:  DealDamageToTheTarget(2),
		}},
	})
}
