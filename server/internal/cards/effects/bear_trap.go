package effects

// Bear Trap — Artifact {1} (Duskmourn: House of Horror, #4466):
//
//	"Flash
//	 {3}, {T}, Sacrifice this artifact: It deals 3 damage to target
//	 creature."
//
// Mouser Foundry's sacrifice-for-damage row with a tap added to the
// cost. The sacrificed Trap deals the damage on its last-known
// information (CR 113.7a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f7afb2c3-50ee-40f5-8b14-32f80c2bacd0",
		Name:            "Bear Trap",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash"},
		Activated: []ActivatedAbility{{
			Label:   "{3}, {T}, Sacrifice this artifact: It deals 3 damage to target creature.",
			Cost:    Plus(ManaCost("{3}"), TapCost(), SacrificeThis()),
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(DamageToTarget(0, 3)),
			Effect:  sourceDealsDamageToEachLegalTarget(3),
		}},
	})
}
