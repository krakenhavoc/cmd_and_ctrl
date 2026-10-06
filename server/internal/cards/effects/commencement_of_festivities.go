package effects

// Commencement of Festivities — Instant {1}{G}:
//
//	"Prevent all combat damage that would be dealt to players this
//	 turn."
//
// #2045: every player, and no permanent, combat damage only. Creatures,
// planeswalkers and battles still take their combat damage.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "56ea8f6b-a765-44e7-bf0e-4d4a2405684b",
		Name:         "Commencement of Festivities",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(combatDamageToPlayersShield()),
	})
}
