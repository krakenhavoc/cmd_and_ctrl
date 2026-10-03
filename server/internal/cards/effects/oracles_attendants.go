package effects

// Oracle's Attendants — Creature — Human Soldier {3}{W}, 1/5:
//
//	"{T}: All damage that would be dealt to target creature this turn by a
//	 source of your choice is dealt to this creature instead."
//
// ADR 0108 §9 (#1905): a redirection for the rest of the turn from a
// source chosen as the ability resolves (CR 609.7a), from the target to
// the Attendants as they are now.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f4bc6674-8586-4dd9-ad3a-87ba704fda7a",
		Name:         "Oracle's Attendants",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{T}: All damage that would be dealt to target creature this turn by a source of your choice is dealt to this creature instead.",
				TapCost(), TargetCreature("target creature"),
				RedirectDamage{Choose: true, Protect: ShieldTheTarget, To: RedirectToThis}),
		},
	})
}
