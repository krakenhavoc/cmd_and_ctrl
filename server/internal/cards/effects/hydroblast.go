package effects

// Hydroblast — Instant {U} (Ice Age, #5636):
//
//	"Choose one —
//	 • Counter target spell if it's red.
//	 • Destroy target permanent if it's red."
//
// Pyroblast's mirror. "If it's red" is checked as the spell resolves,
// not when it is cast, so Hydroblast can target any spell or permanent
// and does nothing to one that isn't red by then; Blue Elemental Blast,
// whose clauses say "target red spell", can't. The body is the shared
// counterOrDestroyTheTargetIfItIs, reading "red" through the spell's
// own text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b94c0336-2ca8-4bb1-bb1f-e9babd1826ad",
		Name:         "Hydroblast",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			Mode("Counter target spell if it's red.", TargetSpell("target spell")),
			Mode("Destroy target permanent if it's red.", TargetPermanent("target permanent")),
		),
		OnResolve: counterOrDestroyTheTargetIfItIs("R"),
	})
}
