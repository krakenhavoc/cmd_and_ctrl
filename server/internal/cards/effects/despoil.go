package effects

// Despoil — Sorcery {3}{B} (Prophecy, #5617):
//
//	"Destroy target land. Its controller loses 2 life."
//
// Stone Rain's target clause with Bitter Downfall's body
// (jmpB2DestroyTargetCreatureItsControllerLosesTwo, which destroys
// whatever its target is: the "creature" is only Bitter Downfall's
// clause). The land's controller is read before the destroy, so the
// life loss still lands when the land is indestructible or regenerated,
// and the whole spell does nothing when its target has left in response
// (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "77fa18fa-5dba-4e4c-9b24-9bfa9b68779f",
		Name:         "Despoil",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target land", Land()),
		OnResolve:    jmpB2DestroyTargetCreatureItsControllerLosesTwo,
	})
}
