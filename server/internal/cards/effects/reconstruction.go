package effects

// Reconstruction — Sorcery {U} (Antiquities, #5640):
//
//	"Return target artifact card from your graveyard to your hand."
//
// Regrowth narrowed to artifact cards: the same graveyard clause Buried
// Ruin prints, picked in the zone browser at announce and re-checked at
// resolution, and Regrowth's body. A card that left the graveyard in
// response leaves the spell to do nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ad8fb78b-5ca5-4ef1-8c68-ee57d1e32fec",
		Name:         "Reconstruction",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target artifact card in your graveyard", Artifact(), YouOwn()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
