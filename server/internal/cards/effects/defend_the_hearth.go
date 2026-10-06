package effects

// Defend the Hearth — Instant {1}{G}:
//
//	"Prevent all combat damage that would be dealt to players this
//	 turn."
//
// #2045: Commencement of Festivities' shield — every player, and no
// permanent, combat damage only.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "1dd64972-5713-4c25-b7a5-a35a56886ef4",
		Name:         "Defend the Hearth",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(combatDamageToPlayersShield()),
	})
}
