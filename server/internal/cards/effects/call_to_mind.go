package effects

// Call to Mind — Sorcery {2}{U} (Magic 2011, #5579):
//
//	"Return target instant or sorcery card from your graveyard to your
//	 hand."
//
// Relearn's text and body: a target that left the graveyard in response
// does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72e2b883-a284-4ab6-8174-1ca289141a49",
		Name:         "Call to Mind",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target instant or sorcery card from your graveyard", Or(Instant(), Sorcery()), YouOwn()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
