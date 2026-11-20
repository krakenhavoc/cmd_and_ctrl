package effects

// Craterize — Sorcery {3}{R} (Magic 2013, #5568):
//
//	"Destroy target land."
//
// Sinkhole in red: one target clause and the shared "Destroy target
// [permanent]." body.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2825f7b0-21a9-44e2-b81f-542686981107",
		Name:         "Craterize",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    destroyTheTargetPermanent,
	})
}
