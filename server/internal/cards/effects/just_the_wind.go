package effects

// Just the Wind — Instant {1}{U} (Shadows over Innistrad, #5543):
//
//	"Return target creature to its owner's hand.
//	 Madness {U} (If you discard this card, discard it into exile. When
//	 you do, cast it for its madness cost or put it into your graveyard.)"
//
// Void Squall's bounce (bounceTheTarget) on a creature. Madness is grown
// by the engine from the one string (CR 702.35a, #657).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "761ebd6a-8f53-46fa-b3cd-16616b773007",
		Name:         "Just the Wind",
		Completeness: CompletenessFull,
		Madness:      "{U}",
		Targets:      TargetCreature("target creature"),
		OnResolve:    bounceTheTarget,
	})
}
