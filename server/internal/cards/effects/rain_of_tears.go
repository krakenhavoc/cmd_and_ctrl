package effects

// Rain of Tears — Sorcery {1}{B}{B} (Portal, #5630):
//
//	"Destroy target land."
//
// Stone Rain's shape (destroyTheTargetPermanent).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "72cecab3-519e-4a23-9623-b423a5c5a251",
		Name:         "Rain of Tears",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
