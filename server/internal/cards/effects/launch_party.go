package effects

// Launch Party — Instant {3}{B} (Jumpstart, #4892):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Destroy target creature. Its controller loses 2 life."
//
// The sacrifice is paid at CR 601.2h with the spell already on the stack,
// so the sacrificed creature cannot usefully be the target: it is gone
// before the spell resolves and the target is illegal (CR 608.2b). The
// body is Bitter Downfall's: the controller is read before the destroy,
// so an indestructible creature's controller still loses 2 life.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "f713a6b4-bd9c-4bc3-9e00-26ecf16e4a92",
		Name:           "Launch Party",
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		Targets:        TargetCreature("target creature"),
		OnResolve:      jmpB2DestroyTargetCreatureItsControllerLosesTwo,
	})
}
