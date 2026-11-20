package effects

// Daring Demolition — Sorcery {2}{B}{B} (Aether Revolt,
// #5536):
//
//	"Destroy target creature or Vehicle."
//
// Spin Out at sorcery speed: an uncrewed Vehicle is a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "8c8a3962-42a5-4070-8719-00d7483c7dbe",
		Name:         "Daring Demolition",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or Vehicle", Or(Creature(), HasSubtype("Vehicle"))),
		OnResolve:    destroyTheTargetPermanent,
	})
}
