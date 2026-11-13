package effects

// Shorecomber Crab — Creature — Crab {U}, 0/4 (Foundations Jumpstart,
// #5464):
//
//	(no rules text)
//
// A vanilla body, registered so the catalog knows it (see bear_cub.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6c8f4d50-c8a6-4ee7-bef2-d91fb595c384",
		Name:         "Shorecomber Crab",
		Completeness: CompletenessFull,
	})
}
