package effects

// Hideous End — Instant {1}{B}{B} (Planechase, #5584):
//
//	"Destroy target nonblack creature. Its controller loses 2 life."
//
// Doom Blade's target with Bitter Downfall's body: the creature's
// controller is read before the destroy, so the life loss still lands when
// the creature is indestructible or regenerated, and the whole spell does
// nothing when its target has left or turned black in response
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f1eafb69-e7c0-4a22-9e73-6bc32168a7c5",
		Name:         "Hideous End",
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target nonblack creature", NonBlack()),
		OnResolve:    jmpB2DestroyTargetCreatureItsControllerLosesTwo,
	})
}
