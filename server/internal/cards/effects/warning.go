package effects

// Warning — Instant {W}:
//
//	"Prevent all combat damage that would be dealt by target attacking creature this turn."
//
// ADR 0108 §7 (#1904, Delivery PR 7b): the shield's source is the
// targeted attacker, pinned as the spell resolves (CR 400.7).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "c5ba0f0f-65c5-4ffa-987a-f320b401ec8f",
		Name:         "Warning",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target attacking creature", AttackingCreature()),
		OnResolve:    shieldAgainstTargetsSpell(true),
	})
}
