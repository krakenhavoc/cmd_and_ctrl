package effects

// Forfend — Instant {1}{W}:
//
//	"Prevent all damage that would be dealt to creatures this turn."
//
// #2045: the not-one-use shield with no source, protecting every
// creature, whoever controls it, and no player. The set is read as the
// damage would be dealt (CR 611.2c): a creature that enters after
// Forfend resolved is protected, and a permanent that stops being a
// creature is not.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "59f8f2c4-d30d-411c-9fac-a5a3219bb124",
		Name:         "Forfend",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldCreatures}),
	})
}
