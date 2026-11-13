package effects

// Eastern Paladin — Creature — Phyrexian Zombie Knight {2}{B}{B}, 3/3
// (Urza's Saga):
//
//	"{B}{B}, {T}: Destroy target green creature."
//
// Exorcist's row with a green target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b445482a-9422-41c7-bedb-3b9f500d2f27",
		Name:         "Eastern Paladin",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{B}{B}, {T}: Destroy target green creature.",
			Cost:    Plus(ManaCost("{B}{B}"), TapCost()),
			Targets: TargetCreature("target green creature", OfColor("G")),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
