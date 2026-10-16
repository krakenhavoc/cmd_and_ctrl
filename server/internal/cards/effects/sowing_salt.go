package effects

// Sowing Salt — Sorcery {2}{R}{R}:
//
//	"Exile target nonbasic land. Search its controller's graveyard, hand,
//	 and library for all cards with the same name as that land and exile
//	 them. Then that player shuffles."
//
// Eradicate's body on a nonbasic land (exilePermanentThenExtract, #4033):
// the name and controller are the land's as it last existed on the
// battlefield; every card with that name in the controller's graveyard is
// exiled (a public zone has no fail to find, CR 701.23b), the caster may
// leave matches in the hand and the library, and the library is shuffled
// whatever was taken. "Nonbasic" is the Basic supertype, so a nonbasic
// land given a basic land type is still a legal target. The search
// happens even if the land did not reach exile; a target that is no
// longer legal makes the spell do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "131d2b70-959d-4814-b0c8-38031f141718",
		Name:         "Sowing Salt",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target nonbasic land", b10NonbasicLand()),
		OnResolve:    exilePermanentThenExtract("Sowing Salt"),
	})
}
