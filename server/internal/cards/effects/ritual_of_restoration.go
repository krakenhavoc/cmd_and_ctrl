package effects

// Ritual of Restoration — Sorcery {W} (Darksteel, #5605):
//
//	"Return target artifact card from your graveyard to your hand."
//
// Reconstruction's regrowth of an artifact card, the target re-checked
// as the spell resolves (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "7cfd0320-0b1c-4ef4-8e17-600cbcbddd6f",
		Name:         "Ritual of Restoration",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target artifact card in your graveyard", Artifact(), YouOwn()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
