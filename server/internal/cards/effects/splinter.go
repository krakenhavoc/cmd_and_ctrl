package effects

// Splinter — Sorcery {2}{G}{G}:
//
//	"Exile target artifact. Search its controller's graveyard, hand, and
//	 library for all cards with the same name as that artifact and exile
//	 them. Then that player shuffles."
//
// Eradicate's body on an artifact (exilePermanentThenExtract, #4033): the
// name and controller are the artifact's as it last existed on the
// battlefield; every card with that name in the controller's graveyard is
// exiled (a public zone has no fail to find, CR 701.23b), the caster may
// leave matches in the hand and the library, and the library is shuffled
// whatever was taken. The search happens even if the artifact did not
// reach exile; a target that is no longer legal makes the spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "90f01e9f-68f1-4e41-97dc-8de990a91504",
		Name:         "Splinter",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve:    exilePermanentThenExtract("Splinter"),
	})
}
