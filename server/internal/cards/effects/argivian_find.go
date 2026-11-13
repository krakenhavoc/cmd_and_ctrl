package effects

// Argivian Find — Instant {W} (Weatherlight, #5629):
//
//	"Return target artifact or enchantment card from your graveyard to your
//	 hand."
//
// Elven Cache's return, narrowed to an artifact or enchantment card. A
// target that left the graveyard in response does nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0b6c74a7-d7d3-4c05-84b1-bcb580bb476f",
		Name:         "Argivian Find",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target artifact or enchantment card from your graveyard", Or(Artifact(), Enchantment()), YouOwn()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
