package effects

// Swooping Lookout — Artifact Creature — Phyrexian Construct {W}, 1/2 (Phyrexia: All Will Be One, #5479):
//
//	"Flying, vigilance"
//
// Every printed keyword is read through the layered view. Toxic is the
// engine's damage tail (infect_wither_toxic.go, CR 702.164).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "0594f19c-c4c2-4812-8f4d-5630fc24c47a",
		Name:            "Swooping Lookout",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "vigilance"},
	})
}
