package effects

// Relearn — Sorcery {1}{U}{U} (Weatherlight, #5629):
//
//	"Return target instant or sorcery card from your graveyard to your hand."
//
// Elven Cache's return, narrowed to an instant or sorcery card. A target
// that left the graveyard in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "13cb7d87-ff12-4742-8f54-f3175536db40",
		Name:         "Relearn",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target instant or sorcery card from your graveyard", Or(Instant(), Sorcery()), YouOwn()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
