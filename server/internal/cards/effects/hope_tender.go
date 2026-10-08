package effects

// Hope Tender — Creature — Human Druid {1}{G}, 2/2:
//
//	"{1}, {T}: Untap target land.
//	 {1}, {T}, Exert this creature: Untap two target lands. (An exerted
//	 creature won't untap during your next untap step.)"
//
// Both untap any land, yours or not. The second needs two different
// lands to target and pays an exert as well (ADR 0130 §4, CR 701.43a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6f377da9-7d7b-4407-8846-75a12b4dd72d",
		Name:         "Hope Tender",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}: Untap target land.",
				Cost:    Plus(ManaCost("{1}"), TapCost()),
				Targets: TargetPermanent("target land", Land()),
				Effect:  b22UntapEachLegalTarget,
			},
			{
				Label:   "{1}, {T}, Exert this creature: Untap two target lands.",
				Cost:    Plus(ManaCost("{1}"), TapCost(), ExertThis()),
				Targets: TargetPermanent("two target lands", Land()).WithCount(2, 2),
				Effect:  b22UntapEachLegalTarget,
			},
		},
	})
}
