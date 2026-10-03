package effects

// Reflect Damage — Instant {3}{R}{W}:
//
//	"The next time a source of your choice would deal damage this turn,
//	 that damage is dealt to that source's controller instead."
//
// ADR 0108 §9 (#1905): a "next time" redirection from a source chosen as
// it resolves (CR 609.7a), whatever it would deal the damage to, to that
// source's controller as the source is when the damage would be dealt.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "21565b0d-f814-49ca-8613-f40040c4ba6c",
		Name:         "Reflect Damage",
		Completeness: CompletenessFull,
		OnResolve:    redirectSpell(RedirectDamage{Choose: true, Protect: ShieldAnything, Next: true, To: RedirectToSourceController}),
	})
}
