package effects

// Brace for Impact — Instant {4}{W}:
//
//	"Prevent all damage that would be dealt to target multicolored
//	 creature this turn. For each 1 damage prevented this way, put a
//	 +1/+1 counter on that creature."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the not-one-use shield pinned to
// the target, with a CR 615.5 follow-up. The follow-up runs once for each
// instance of damage the shield prevents, immediately after it, with
// that instance's total (CR 615.13's "each time a prevention effect is
// applied"), and puts that many counters on the creature. Damage that
// can't be prevented is dealt and gives no counters (CR 615.12).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "f250fb94-e519-4bfa-921e-eddeeafcf76c",
		Name:         "Brace for Impact",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target multicolored creature", Multicolored()),
		OnResolve: sourceShieldSpell(PreventDamageFromSource{
			Protect: ShieldTheTarget,
			Then:    preventedPlusOneCountersOnRecipientBody,
		}),
	})
}
