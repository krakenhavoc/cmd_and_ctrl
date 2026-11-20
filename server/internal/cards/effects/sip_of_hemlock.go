package effects

// Sip of Hemlock — Sorcery {4}{B}{B} (Theros, #5561):
//
//	"Destroy target creature. Its controller loses 2 life."
//
// Bitter Downfall's body: the creature is destroyed and its controller
// loses 2 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "861096be-dbb3-429e-820f-d82f6ed2aa24",
		Name:         "Sip of Hemlock",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve:    jmpB2DestroyTargetCreatureItsControllerLosesTwo,
	})
}
