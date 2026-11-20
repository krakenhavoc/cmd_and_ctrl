package effects

// Temporal Eddy — Sorcery {2}{U}{U} (Time Spiral, #5596):
//
//	"Put target creature or land on top of its owner's library."
//
// Submerge's tuck over a creature or a land (putTargetOnTopOfOwnersLibrary).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3a35e772-2547-41a3-a2a2-125b48c23aa5",
		Name:         "Temporal Eddy",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target creature or land", Or(Creature(), Land())),
		OnResolve:    putTargetOnTopOfOwnersLibrary,
	})
}
