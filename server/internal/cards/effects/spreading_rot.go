package effects

// Spreading Rot — Sorcery {4}{B} (Ixalan, #5532):
//
//	"Destroy target land. Its controller loses 2 life."
//
// Despoil's text: the land's controller as it is destroyed loses the
// life (its last-known controller, CR 608.2h).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3c1b95bd-a2fc-47cd-8940-e4af03bd3ce2",
		Name:         "Spreading Rot",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    jmpB2DestroyTargetCreatureItsControllerLosesTwo,
	})
}
