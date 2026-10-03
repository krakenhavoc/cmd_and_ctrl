package effects

// Endure — Instant {3}{W}{W}:
//
//	"Prevent all damage that would be dealt to you and permanents you
//	 control this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the not-one-use shield with no
// source, protecting you and every permanent you control, read as the
// damage would be dealt: a permanent that comes under your control later
// this turn is protected too, and one you lose control of is not.
// (CR 611.2c fixes the affected set only for an effect that changes
// characteristics or control; a prevention effect does neither.)
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f4cd32f0-d6aa-4497-b8db-a0bf4c3c31de",
		Name:         "Endure",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldYouAndPermanentsYouControl}),
	})
}
