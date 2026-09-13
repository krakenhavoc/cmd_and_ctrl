package effects

// Ioreth of the Healing House — Legendary Creature — Human Cleric
// {2}{U}, 1/4 (EDHREC rank 2430):
//
//	"{T}: Untap another target permanent.
//	 {T}: Untap two other target legendary creatures."
//
// Kiora's Follower with a second, wider mode. Two CR 602 activated
// abilities sharing the tap (so one activation per untap of Ioreth);
// the first untaps any one other permanent, anyone's; the second is
// a two-slot target clause over other legendary creatures — the
// legend-heavy deck's double untap. "Another / other" is Kiora's
// Follower's b03NotNamed posture: the picker cannot see the source,
// so Ioreth is excluded by name, which also excludes an opponent's
// Ioreth — a permanent the legend rule keeps rare and one the
// controller would seldom want to untap. Both slots of the second
// ability must be filled; with one legendary creature on the
// battlefield it cannot be activated.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "af6e4c3e-0276-4f72-9a70-25868fe8bba5",
		Name:         "Ioreth of the Healing House",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Untap another target permanent.",
				Cost:    TapCost(),
				Targets: TargetPermanent("another target permanent", b03NotNamed("Ioreth of the Healing House")),
				Effect:  b22UntapEachLegalTarget,
			},
			{
				Label: "{T}: Untap two other target legendary creatures.",
				Cost:  TapCost(),
				Targets: TargetCreature("two other target legendary creatures",
					b05Legendary(), b03NotNamed("Ioreth of the Healing House")).WithCount(2, 2),
				Effect: b22UntapEachLegalTarget,
			},
		},
	})
}
