package effects

// Fleetfoot Dancer (Streets of New Capenna, #5491):
//
//	Trample, lifelink, haste
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "8f606acf-db16-42cd-8907-dad88c4a1d54",
		Name:            "Fleetfoot Dancer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample", "lifelink", "haste"},
	})
}
