package effects

// Razia, Boros Archangel — Legendary Creature — Angel {4}{R}{R}{W}{W}, 6/3:
//
//	"Flying, vigilance, haste
//	 {T}: The next 3 damage that would be dealt to target creature you
//	 control this turn is dealt to another target creature instead."
//
// ADR 0108 §9 (#1905): a 3-point charged redirection (CR 615.7) from the
// first target to the second. The rulings: the 3 need not come from one
// source or at once; either target leaving stops it; Razia leaving does
// not.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "d830a136-6fb9-42e7-81f1-97e1d713ee82",
		Name:         "Razia, Boros Archangel",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			redirectRow("{T}: The next 3 damage that would be dealt to target creature you control this turn is dealt to another target creature instead.",
				TapCost(), Clauses(
					TargetCreature("target creature you control", YouControl()),
					Distinct(TargetCreature("another target creature")),
				),
				RedirectDamage{Protect: ShieldClause(0), Amount: 3, To: RedirectToClause(1)}),
		},
	})
}
