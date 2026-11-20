package effects

// Darksteel Relic — Artifact {0} (New Phyrexia, #5575):
//
//	"Indestructible (Effects that say "destroy" don't destroy this artifact.)"
//
// A keyword permanent, Darksteel Myr's: indestructible is the engine's
// (CR 702.12b), on any permanent.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "2bda1fb1-528c-4e66-89f1-bb1a85e9181e",
		Name:            "Darksteel Relic",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"indestructible"},
	})
}
