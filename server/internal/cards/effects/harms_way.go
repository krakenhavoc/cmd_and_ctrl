package effects

// Harm's Way — Instant {W}:
//
//	"The next 2 damage that a source of your choice would deal to you
//	 and/or permanents you control this turn is dealt to any target
//	 instead."
//
// ADR 0108 §9 and owner decision 1 (#1905): a charged redirection
// (CR 615.7) from a source chosen as it resolves (CR 609.7a) to the
// target, chosen as it was cast. The rulings: 1 damage redirected leaves
// a 1-point shield for later in the turn; damage to several of your
// things at once is divided — you choose which 2 is redirected
// (divide_shield); a target that can't be dealt damage when the damage
// would be redirected leaves it where it was.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "4c34a882-3786-4d3c-9ca4-04fa85d5e51b",
		Name:         "Harm's Way",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		OnResolve:    redirectSpell(RedirectDamage{Choose: true, Protect: ShieldYouAndPermanentsYouControl, Amount: 2, To: RedirectToClause(0)}),
	})
}
