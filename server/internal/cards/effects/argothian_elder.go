package effects

// Argothian Elder — Creature — Elf Druid {3}{G}, 2/2 (Urza's Saga):
//
//	"{T}: Untap two target lands."
//
// Ley Weaver's ability without the partner: exactly two target lands,
// any player's. A land that has left by resolution is skipped
// (CR 608.2b), and the {T} cost waits out summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b57f06e4-da97-4473-9fbf-f66e9129502a",
		Name:         "Argothian Elder",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Untap two target lands.",
			Cost:    TapCost(),
			Targets: TargetPermanent("two target lands", Land()).WithCount(2, 2),
			Effect:  UntapEachLegalTarget,
		}},
	})
}
