package effects

// Elven Cache — Sorcery {2}{G}{G} (Visions, #5631):
//
//	"Return target card from your graveyard to your hand."
//
// Regrowth's text at a higher cost, and Regrowth's body
// (returnTargetFromGraveyardToHand, regrowth.go). A target that left the
// graveyard in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8a922366-ee2a-4ddb-a904-683dcf4f191a",
		Name:         "Elven Cache",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card from your graveyard", YouOwn()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
