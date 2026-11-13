package effects

// Winter's Grasp — Sorcery {1}{G}{G} (Portal, #5630):
//
//	"Destroy target land."
//
// Stone Rain's shape (destroyTheTargetPermanent).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e9b8679d-52a9-4f0f-9365-f3e4b7a69805",
		Name:         "Winter's Grasp",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
