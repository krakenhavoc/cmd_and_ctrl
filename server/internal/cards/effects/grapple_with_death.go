package effects

// Grapple with Death — Sorcery {1}{B}{G} (Secrets of Strixhaven, #4456):
//
//	"Destroy target artifact or creature. You gain 1 life."
//
// The life is its own sentence, so you gain it even when the destroy did
// nothing (indestructible, a regeneration shield, a replacement); a spell
// whose only target has left does nothing at all (CR 608.2b). Terminal
// Criticism's body, shared (socB26DestroyTheTargetAndGainLife).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dced5cab-1438-4264-b44c-54dc37ec73d3",
		Name:         "Grapple with Death",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact or creature", Or(Artifact(), Creature())),
		OnResolve:    socB26DestroyTheTargetAndGainLife(1),
	})
}
