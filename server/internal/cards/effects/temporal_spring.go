package effects

// Temporal Spring — Sorcery {1}{G}{U} (Apocalypse, #5614):
//
//	"Put target permanent on top of its owner's library."
//
// Submerge's tuck widened to any permanent (putTargetOnTopOfOwnersLibrary):
// it goes on top of its OWNER's library whoever controls it, and a token
// put there ceases to exist (CR 111.8). A target gone by resolution
// leaves the spell nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1d7a4e9c-b3c9-421e-b5f4-82d51d7d5125",
		Name:         "Temporal Spring",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target permanent"),
		OnResolve:    putTargetOnTopOfOwnersLibrary,
	})
}
