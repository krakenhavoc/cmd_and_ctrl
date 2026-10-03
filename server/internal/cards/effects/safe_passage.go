package effects

// Safe Passage — Instant {2}{W}:
//
//	"Prevent all damage that would be dealt to you and creatures you
//	 control this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the not-one-use shield with no
// source, protecting you and the creatures you control as the damage
// would be dealt (Endure's reading, narrowed to creatures).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "dfa459a1-b065-4488-88d3-4da388261b52",
		Name:         "Safe Passage",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldYouAndYourCreatures}),
	})
}
