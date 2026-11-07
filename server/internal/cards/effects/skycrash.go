package effects

// Skycrash — Instant {1}{R} (Aetherdrift, #4464):
//
//	"Destroy target artifact.
//	 Cycling {R} ({R}, Discard this card: Draw a card.)"
//
// The destroy is destroyTheTargetPermanent on any artifact, a Vehicle
// included; cycling is the engine's hand ability (#660).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2343ebff-edf4-4132-ad9e-7800f8014072",
		Name:         "Skycrash",
		Completeness: CompletenessFull,
		Targets:      TargetPermanent("target artifact", Artifact()),
		OnResolve:    destroyTheTargetPermanent,
		Activated:    []ActivatedAbility{Cycling("{R}")},
	})
}
