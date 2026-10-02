package effects

// Reverse Damage — Instant {1}{W}{W}:
//
//	"The next time a source of your choice would deal damage to you this turn, prevent that damage. You gain life equal to the damage prevented this way."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// "You gain life equal to the damage prevented this way" is the shield's
// own additional effect, run immediately after the prevention (CR 615.5,
// owner decision 4).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "eaaf7c30-f463-4115-a40e-7dc717063413",
		Name:         "Reverse Damage",
		Completeness: CompletenessFull,
		OnResolve:    nextDamageShieldSpell(PreventNextDamageFromChosenSource(ShieldYou).WithThen(preventedGainLifeBody)),
	})
}
