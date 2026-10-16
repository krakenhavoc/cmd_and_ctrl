package effects

// Scour — Instant {2}{W}{W}:
//
//	"Exile target enchantment. Search its controller's graveyard, hand,
//	 and library for all cards with the same name as that enchantment and
//	 exile them. Then that player shuffles."
//
// Eradicate's body on an enchantment (exilePermanentThenExtract, #4033):
// the name and controller are the enchantment's as it last existed on the
// battlefield; every card with that name in the controller's graveyard is
// exiled (a public zone has no fail to find, CR 701.23b), the caster may
// leave matches in the hand and the library, and the library is shuffled
// whatever was taken. The search happens even if the enchantment did not
// reach exile; a target that is no longer legal makes the spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f18622da-81f3-4854-a1bb-8d658751d8db",
		Name:         "Scour",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve:    exilePermanentThenExtract("Scour"),
	})
}
