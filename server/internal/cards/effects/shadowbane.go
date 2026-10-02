package effects

// Shadowbane — Instant {1}{W}:
//
//	"The next time a source of your choice would deal damage to you and/or creatures you control this turn, prevent that damage. If damage from a black source is prevented this way, you gain that much life."
//
// ADR 0107 §6 (#1860): the one-use shield against the next instance of
// damage from a source chosen as it resolves (CR 615.8,
// 609.7a).
// It protects you and the creatures you control, judged as the damage
// would be dealt. "If damage from a black source is prevented this way,
// you gain that much life" is the shield's additional effect (CR 615.5).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "9e9ec33b-b923-40e5-81a8-e1e78df63ec7",
		Name:         "Shadowbane",
		Completeness: CompletenessFull,
		OnResolve:    nextDamageShieldSpell(PreventNextDamageFromChosenSource(ShieldYouAndYourCreatures).WithThen(preventedBlackGainLifeBody)),
	})
}
