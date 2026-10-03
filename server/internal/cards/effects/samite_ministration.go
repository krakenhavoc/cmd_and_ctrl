package effects

// Samite Ministration — Instant {1}{W}:
//
//	"Prevent all damage that would be dealt to you this turn by a source of your choice. Whenever damage from a black or red source is prevented this way this turn, you gain that much life."
//
// ADR 0108 §7 decisions 1 and 5 (#1904): a shield against a source chosen
// as it resolves (CR 609.7a) that protects you from every instance of its
// damage this turn. "Whenever damage … is prevented this way" is a
// triggered ability (CR 615.13): it triggers once each time the shield is
// applied to damage dealt at the same time and prevents some of it, and
// goes on the stack with the amount. The source's colour is read as it
// was when the damage would have been dealt, so a source that is black
// or red then gains you life even if it changes colour afterwards.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "af669df1-3526-46ca-b90c-19b112b7ba44",
		Name:         "Samite Ministration",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(PreventDamageFromChosenSource(ShieldYou).WithThen(preventedBlackOrRedTriggerBody)),
	})
}
