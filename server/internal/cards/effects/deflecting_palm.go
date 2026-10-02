package effects

// Deflecting Palm — Instant {R}{W}:
//
//	"The next time a source of your choice would deal damage to you this turn, prevent that damage. If damage is prevented this way, Deflecting Palm deals that much damage to that source's controller."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// "If damage is prevented this way, Deflecting Palm deals that much
// damage to that source's controller" is the shield's additional effect,
// run immediately after the prevention (CR 615.5, owner decision 4). The
// controller is the source's as it dealt the damage; the damage comes
// from Deflecting Palm, red and white, as it last existed.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "dc5dffc8-fac5-4956-bac4-1ad2cc16f6be",
		Name:         "Deflecting Palm",
		Completeness: CompletenessFull,
		OnResolve:    nextDamageShieldSpell(PreventNextDamageFromChosenSource(ShieldYou).WithThen(preventedDamageSourceControllerBody)),
	})
}
