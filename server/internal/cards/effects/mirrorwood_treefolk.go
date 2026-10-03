package effects

// Mirrorwood Treefolk — Creature — Treefolk {3}{G}, 2/4:
//
//	"{2}{R}{W}: The next time damage would be dealt to this creature this
//	 turn, that damage is dealt to any target instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection with no source, from
// the Treefolk to the target. The rulings: a target that can't be dealt
// damage leaves the damage on the Treefolk; several sources dealing it
// damage at once (in combat) are one instance, and all of it is
// redirected (CR 615.8).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "fec51ab9-484f-46a8-b2b0-772a61d89e41",
		Name:         "Mirrorwood Treefolk",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{2}{R}{W}: The next time damage would be dealt to this creature this turn, that damage is dealt to any target instead.",
				ManaCost("{2}{R}{W}"), TargetAny(),
				RedirectDamage{Protect: ShieldThis, Next: true, To: RedirectToClause(0)}),
		},
	})
}
