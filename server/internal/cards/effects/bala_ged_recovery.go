package effects

// Bala Ged Recovery // Bala Ged Sanctuary — Modal double-faced card
// (EDHREC top 1,000, #3654). The front face:
//
//	"Return target card from your graveyard to your hand." {2}{G} Sorcery
//
// Regrowth on the front of an MDFC. The land back face is the shared
// mdfc_lands.go cycle, registered under "<oracle_id>#1".
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2075f58-b0e9-4e85-b7e6-0523a27a1d5b",
		Name:         "Bala Ged Recovery",
		Completeness: CompletenessFull,
		Targets:      TargetCardInGraveyard("target card in your graveyard", YouOwn()),
		OnResolve:    returnTargetFromGraveyardToHand,
	})
}
