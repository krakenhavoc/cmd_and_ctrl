package effects

// Redeem — Instant {1}{W}:
//
//	"Prevent all damage that would be dealt this turn to up to two target
//	 creatures."
//
// ADR 0108 §7, Delivery PR 7 (#1904): one prevention effect, so one
// not-one-use record protecting every chosen creature
// (ShieldTheTargetPermanents). A target gone by resolution is not
// protected, and the others still are (CR 608.2b).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "c2032f88-a000-4ed2-a7e5-61d29723d45e",
		Name:         "Redeem",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldTheTargetPermanents}),
	})
}
