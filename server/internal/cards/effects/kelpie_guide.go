package effects

// Kelpie Guide — Creature — Beast {2}{U}, 2/2 (EDHREC rank 4606):
//
//	"{T}: Untap another target permanent you control.
//	 {T}: Tap target permanent. Activate only if you control eight or
//	 more lands."
//
// An untapper that turns into an Icy Manipulator late. The second
// ability's gate is its activation condition (CR 602.1b, #743),
// ControlsAtLeast(8, lands). Both are {T} abilities on a creature, so
// summoning sickness applies to both (CR 302.6).
//
// One declared simplification, weaker than printed: "another target
// permanent you control" is matched by NAME, the catalog's convention
// for "another" on an activated ability's target clause (Benevolent
// Hydra, Ioreth) — a target predicate is not told which permanent is
// activating. The Guide cannot untap itself, as printed, and cannot
// untap a second Kelpie Guide either, which the printed card can.
func init() {
	Register(Spec{
		OracleID:     "8fc1e8b1-3afc-4e9d-a3ae-7bd9bcdcb465",
		Name:         "Kelpie Guide",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The untap ability can't untap another Kelpie Guide."},
		Activated: []ActivatedAbility{
			{
				Label:   "{T}: Untap another target permanent you control.",
				Cost:    TapCost(),
				Targets: TargetPermanent("another target permanent you control", YouControl(), b03NotNamed("Kelpie Guide")),
				Effect:  b22UntapEachLegalTarget,
			},
			{
				Label:     "{T}: Tap target permanent. Activate only if you control eight or more lands.",
				Cost:      TapCost(),
				Targets:   TargetPermanent("target permanent", Permanent()),
				Condition: ControlsAtLeast(8, MatchLand),
				Effect:    b36TapChosenCreature,
			},
		},
	})
}
