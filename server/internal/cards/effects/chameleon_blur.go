package effects

// Chameleon Blur — Instant {3}{G}:
//
//	"Prevent all damage that creatures would deal to players this
//	 turn."
//
// #2045 beside ADR 0108 §7's property recheck: the source is a creature
// as it would deal the damage (CR 615.9), and the recipient is a player.
// Combat damage or not. Damage a creature deals to a creature or a
// planeswalker is dealt, and so is damage a noncreature source deals to
// a player.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "6ac8ac31-02ce-48e0-98b9-1c5858e1f593",
		Name:         "Chameleon Blur",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromSource{Protect: ShieldPlayers, Queries: creatureSources()}),
	})
}
